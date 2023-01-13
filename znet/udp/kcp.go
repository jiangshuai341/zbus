package udp

import (
	"encoding/binary"
	"github.com/jiangshuai341/zbus/zpool"
	"sync/atomic"
	"time"
)

/*
0               4   5   6       8 (BYTE)
+---------------+---+---+-------+
|     conv      |cmd|frg|  wnd  |
+---------------+---+---+-------+   8
|     ts        |     sn        |
+---------------+---------------+  16
|     una       |     len       |
+---------------+---------------+  24
|                               |
|        DATA (optional)        |
|                               |
+-------------------------------+

conv: 4 字节: 连接标识

cmd: 1 字节: Command.
	IKCP_CMD_PUSH 数据报文
	IKCP_CMD_ACK 确认报文
	IKCP_CMD_WASK 询问对端剩余接收窗口的大小
	IKCP_CMD_WINS 通知对端剩余接收窗口的大小

frg  : 1 字节: 分片数量. 表示随后还有多少个报文属于同一个包
wnd  : 2 字节: 发送方剩余接收窗口的大小.
ts   : 4 字节: 时间戳.
sn   : 4 字节: 报文编号.
una  : 4 字节: 发送方的接收缓冲区中最小还未收到的报文段的编号. 也就是说, 编号比它小的报文段都已全部接收.
len  : 4 字节: 数据段长度.
data : 数据段. 只有数据报文会有这个字段.
*/

const (
	IKCP_RTO_NDL = 30  // no delay min rto
	IKCP_RTO_MIN = 100 // normal min rto
	IKCP_RTO_DEF = 200
	IKCP_RTO_MAX = 60000

	IKCP_CMD_PUSH = 81
	IKCP_CMD_ACK  = 82
	IKCP_CMD_WASK = 83
	IKCP_CMD_WINS = 84

	IKCP_ASK_SEND    = 1
	IKCP_ASK_TELL    = 2
	IKCP_WND_SND     = 32
	IKCP_WND_RCV     = 32
	IKCP_MTU_DEF     = 1400
	IKCP_ACK_FAST    = 3
	IKCP_INTERVAL    = 100
	IKCP_OVERHEAD    = 24
	IKCP_DEADLINK    = 20
	IKCP_THRESH_INIT = 2
	IKCP_THRESH_MIN  = 2
	IKCP_PROBE_INIT  = 7000
	IKCP_PROBE_LIMIT = 120000
	IKCP_SN_OFFSET   = 12
)

var refTime time.Time = time.Now()

func currentMs() uint32 {
	return uint32(time.Since(refTime) / time.Millisecond)
}

type output_callback func(buf []byte, size int)

/* encode 8 bits unsigned int */
func ikcp_encode8u(p []byte, c byte) []byte {
	p[0] = c
	return p[1:]
}

/* decode 8 bits unsigned int */
func ikcp_decode8u(p []byte, c *byte) []byte {
	*c = p[0]
	return p[1:]
}

/* encode 16 bits unsigned int (lsb) */
func ikcp_encode16u(p []byte, w uint16) []byte {
	binary.LittleEndian.PutUint16(p, w)
	return p[2:]
}

/* decode 16 bits unsigned int (lsb) */
func ikcp_decode16u(p []byte, w *uint16) []byte {
	*w = binary.LittleEndian.Uint16(p)
	return p[2:]
}

/* encode 32 bits unsigned int (lsb) */
func ikcp_encode32u(p []byte, l uint32) []byte {
	binary.LittleEndian.PutUint32(p, l)
	return p[4:]
}

/* decode 32 bits unsigned int (lsb) */
func ikcp_decode32u(p []byte, l *uint32) []byte {
	*l = binary.LittleEndian.Uint32(p)
	return p[4:]
}

func _imin_(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}

func _imax_(a, b uint32) uint32 {
	if a >= b {
		return a
	}
	return b
}

func _ibound_(lower, middle, upper uint32) uint32 {
	return _imin_(_imax_(lower, middle), upper)
}

func _itimediff(later, earlier uint32) int32 {
	return (int32)(later - earlier)
}

type segment struct {
	conv     uint32
	cmd      uint8
	frg      uint8
	wnd      uint16
	ts       uint32
	sn       uint32
	una      uint32
	rto      uint32
	xmit     uint32 //该报文传输的次数
	resendts uint32 //超时时间戳
	fastack  uint32 //ACK 失序次数
	acked    uint32
	data     []byte
}

type ackItem struct {
	sn uint32
	ts uint32
}

// encode a segment into buffer
func (seg *segment) encode(ptr []byte) []byte {
	ptr = ikcp_encode32u(ptr, seg.conv)
	ptr = ikcp_encode8u(ptr, seg.cmd)
	ptr = ikcp_encode8u(ptr, seg.frg)
	ptr = ikcp_encode16u(ptr, seg.wnd)
	ptr = ikcp_encode32u(ptr, seg.ts)
	ptr = ikcp_encode32u(ptr, seg.sn)
	ptr = ikcp_encode32u(ptr, seg.una)
	ptr = ikcp_encode32u(ptr, uint32(len(seg.data)))
	atomic.AddUint64(&DefaultSnmp.OutSegs, 1)
	return ptr
}

type KCP struct {
	conv, mtu, mss, state                  uint32
	snd_una, snd_nxt, rcv_nxt              uint32
	ssthresh                               uint32
	rx_rttvar, rx_srtt                     int32
	rx_rto, rx_minrto                      uint32
	snd_wnd, rcv_wnd, rmt_wnd, cwnd, probe uint32
	interval, ts_flush                     uint32
	nodelay, updated                       uint32
	ts_probe, probe_wait                   uint32
	dead_link, incr                        uint32

	fastresend     int32
	nocwnd, stream int32

	snd_queue []segment
	rcv_queue []segment
	snd_buf   []segment
	rcv_buf   []segment

	acklist []ackItem

	buffer   []byte
	reserved int
	ouput    output_callback
}

/*
conv							: 连接标识
mtu, mss						: 最大传输单元和最大报文段大小. mss = mtu - 包头长度(24).
state							: 连接状态, 0 表示连接建立, -1 表示连接断开.
snd_una							: 发送缓冲区中最小还未确认送达的报文段的编号.编号比它小的报文段都已确认送达.
snd_nxt							: 下一个等待发送的报文段的编号
rcv_nxt							: 下一个等待接收的报文段的编号.
ts_recent, ts_lastack			: 未使用.
ssthresh						: Slow Start Threshold, 慢启动阈值.
rx_rto							: Retransmission TimeOut(RTO), 超时重传时间.
rx_rttval, rx_srtt, rx_minrto	: 平滑之后rtt,平滑之后的抖动,flash间隔
snd_wnd, rcv_wnd				: 发送窗口和接收窗口的大小.
rmt_wnd							: 对端剩余接收窗口的大小.
cwnd							: congestion window, 拥塞窗口. 用于拥塞控制.
probe							: 是否要发送窗口探测报文的标志.
current							: 当前时间.
interval						: flush 的时间粒度.
ts_flush						: 下次 flush 的时间戳.
xmit							: 该链接超时重传的总次数.
snd_buf, rcv_buf				: 发送缓冲区和接收缓冲区.
snd_queue, rcv_queue			: 发送队列和接收队列.
nodelay							: 是否启动快速模式. 用于控制 RTO 增长速度.
updated							: 是否调用过 ikcp_update.
ts_probe, probe_wait			: 确定何时需要发送窗口询问报文.
dead_link						: 当一个报文发送超时次数达到 dead_link 次时认为连接断开.
incr							: 用于计算 cwnd.
acklist, ackcount, ackblock		: ACK 列表, ACK 列表的长度和容量. 待发送的 ACK 的相关信息会先存在 ACK 列表中, flush 时一并发送.
buffer							: flush 时用到的临时缓冲区.
fastresend						: ACK 失序 fastresend 次时触发快速重传.
fastlimit						: 传输次数小于 fastlimit 的报文才会执行快速重传.
nocwnd							: 是否不考虑拥塞窗口.
stream							: 是否开启流模式, 开启后可能会合并包.
output							: 下层协议输出函数.

*/

/*
发送流程:
(kcp *KCP) Send (Data)
	snd_queue -> snd_buf -> (kcp *KCP) flush(ackOnly bool) : 不将报文直接加入 snd_buf 是为了防止一次发送过多的报文导致拥塞

*/

func NewKCP(conv uint32, output output_callback) *KCP {
	kcp := &KCP{}

	kcp.snd_wnd = IKCP_WND_SND
	kcp.rcv_wnd = IKCP_WND_RCV
	kcp.rmt_wnd = IKCP_WND_RCV
	kcp.mtu = IKCP_MTU_DEF
	kcp.mss = kcp.mtu - IKCP_OVERHEAD
	kcp.buffer = make([]byte, kcp.mtu)
	kcp.rx_rto = IKCP_RTO_DEF
	kcp.rx_minrto = IKCP_RTO_MIN
	kcp.interval = IKCP_INTERVAL
	kcp.ts_flush = IKCP_INTERVAL
	kcp.ssthresh = IKCP_THRESH_INIT
	kcp.dead_link = IKCP_DEADLINK

	kcp.conv = conv
	kcp.ouput = output

	return kcp
}

func (kcp *KCP) newSegment(size int) (seg segment) {
	seg.data = zpool.GetBuffer2(size)
	return
}
func (kcp KCP) delSegment(seg *segment) {
	if seg.data != nil {
		zpool.PutBuffer(seg.data)
	}
}
func (kcp *KCP) remove_front(q []segment, n int) []segment {
	if n > cap(q)/2 {
		newn := copy(q, q[n:])
		return q[:newn]
	}
	return q[n:]
}
func (kcp *KCP) ReserveBytes(n int) bool {
	if n >= int(kcp.mtu-IKCP_OVERHEAD) || n < 0 {
		return false
	}
	kcp.reserved = n
	kcp.mss = kcp.mtu - IKCP_OVERHEAD - uint32(n)

	return true
}
func (kcp *KCP) PeekSize() (length int) {
	if len(kcp.rcv_queue) == 0 {
		return -1
	}
	seg := &kcp.rcv_queue[0]
	if seg.frg == 0 {
		return len(seg.data)
	}
	if len(kcp.rcv_queue) < int(seg.frg+1) {
		return -1
	}
	for k := range kcp.rcv_queue {
		seg := &kcp.rcv_queue[k]
		length += len(seg.data)
		if seg.frg == 0 {
			break
		}
	}
	return
}
func (kcp *KCP) Recv(buffer []byte) (n int) {
	peeksize := kcp.PeekSize()
	if peeksize < 0 {
		return -1
	}
	if peeksize > len(buffer) {
		return -2
	}

	var fast_recover bool
	if len(kcp.rcv_queue) >= int(kcp.rcv_wnd) {
		fast_recover = true
	}
	count := 0
	for k := range kcp.rcv_queue {
		seg := &kcp.rcv_queue[k]
		copy(buffer, seg.data)
		buffer = buffer[len(seg.data):]
		n += len(seg.data)
		count++
		kcp.delSegment(seg)
		if seg.frg == 0 {
			break
		}
	}
	if count > 0 {
		kcp.rcv_queue = kcp.remove_front(kcp.rcv_queue, count)
	}
	count = 0
	for k := range kcp.rcv_buf {
		seg := &kcp.rcv_buf[k]
		if seg.sn == kcp.rcv_nxt &&
			len(kcp.rcv_queue)+count < int(kcp.rcv_wnd) {
			kcp.rcv_nxt++
			count++
		} else {
			break
		}
	}
	if count > 0 {
		kcp.rcv_queue = append(kcp.rcv_queue, kcp.rcv_buf[:count]...)
		kcp.rcv_buf = kcp.remove_front(kcp.rcv_buf, count)
	}

	if len(kcp.rcv_queue) < int(kcp.rcv_wnd) && fast_recover {
		kcp.probe |= IKCP_ASK_TELL
	}
	return
}

func (kcp *KCP) Send(buffer []byte) int {
	var count int
	if len(buffer) == 0 {
		return -1
	}
	if kcp.stream != 0 {
		n := len(kcp.snd_queue)
		if n > 0 {
			seg := &kcp.snd_queue[n-1]
			if len(seg.data) < int(kcp.mss) {
				capacity := int(kcp.mss) - len(seg.data)
				extend := capacity
				if len(buffer) < capacity {
					extend = len(buffer)
				}
				oldlen := len(seg.data)
				seg.data = seg.data[:oldlen+extend]
				copy(seg.data[oldlen:], buffer)
				buffer = buffer[extend:]
			}
		}
		if len(buffer) == 0 {
			return 0
		}
	}
	if len(buffer) <= int(kcp.mss) {
		count = 1
	} else {
		count = (len(buffer) + int(kcp.mss) - 1) / int(kcp.mss)
	}
	if count > 255 {
		return -2
	}
	if count == 0 {
		count = 1
	}
	for i := 0; i < count; i++ {
		var size int
		if len(buffer) > int(kcp.mss) {
			size = int(kcp.mss)
		} else {
			size = len(buffer)
		}
		seg := kcp.newSegment(size)
		copy(seg.data, buffer[:size])
		if kcp.stream == 0 {
			seg.frg = uint8(count - i - 1)
		} else {
			seg.frg = 0
		}
		kcp.snd_queue = append(kcp.snd_queue, seg)
		buffer = buffer[size:]
	}
	return 0
}

func (kcp *KCP) update_ack(rtt int32) {
	var rto uint32
	if kcp.rx_srtt == 0 {
		kcp.rx_srtt = rtt
		kcp.rx_rttvar = rtt >> 1
	} else {
		delta := rtt - kcp.rx_srtt
		kcp.rx_srtt += delta >> 3
		if delta < 0 {
			delta = -delta
		}
		if rtt < kcp.rx_srtt-kcp.rx_rttvar {
			kcp.rx_rttvar += (delta - kcp.rx_rttvar) >> 5
		} else {
			kcp.rx_rttvar += (delta - kcp.rx_rttvar) >> 2
		}
	}
	rto = uint32(kcp.rx_srtt) + _imax_(kcp.interval, uint32(kcp.rx_rttvar)<<2)
	kcp.rx_rto = _ibound_(kcp.rx_minrto, rto, IKCP_RTO_MAX)
}
func (kcp *KCP) shrink_buf() {
	if len(kcp.snd_buf) > 0 {
		seg := &kcp.snd_buf[0]
		kcp.snd_una = seg.sn
	} else {
		kcp.snd_una = kcp.snd_nxt
	}
}
func (kcp *KCP) parse_ack(sn uint32) {
	if _itimediff(sn, kcp.snd_una) < 0 || _itimediff(sn, kcp.snd_nxt) >= 0 {
		return
	}
	for k := range kcp.snd_buf {
		seg := &kcp.snd_buf[k]
		if sn == seg.sn {
			seg.acked = 1
			kcp.delSegment(seg)
			break
		}
		if _itimediff(sn, seg.sn) < 0 {
			break
		}
	}
}
func (kcp *KCP) parse_fastack(sn, ts uint32) {
	if _itimediff(sn, kcp.snd_una) < 0 || _itimediff(sn, kcp.snd_nxt) >= 0 {
		return
	}
	for k := range kcp.snd_buf {
		seg := &kcp.snd_buf[k]
		if _itimediff(sn, seg.sn) < 0 {
			break
		} else if sn != seg.sn && _itimediff(seg.ts, ts) <= 0 {
			seg.fastack++
		}
	}
}
func (kcp *KCP) parse_una(una uint32) int {
	count := 0
	for k := range kcp.snd_buf { //为什么不采用根据SN计算数组下标
		seg := &kcp.snd_buf[k]
		if _itimediff(una, seg.sn) > 0 {
			kcp.delSegment(seg)
			count++
		} else {
			break
		}
	}
	if count > 0 {
		kcp.snd_buf = kcp.remove_front(kcp.snd_buf, count)
	}
	return count
}
func (kcp *KCP) ack_push(sn, ts uint32) {
	kcp.acklist = append(kcp.acklist, ackItem{sn, ts})

}
func (kcp *KCP) parse_data(newseg segment) bool {
	sn := newseg.sn
	if _itimediff(sn, kcp.rcv_nxt+kcp.rcv_wnd) >= 0 ||
		_itimediff(sn, kcp.rcv_nxt) < 0 {
		return true
	}
	n := len(kcp.rcv_buf) - 1
	insert_idx := 0
	repeat := false
	for i := n; i >= 0; i-- {
		seg := &kcp.rcv_buf[i]
		if seg.sn == sn {
			repeat = true
			break
		}
		if _itimediff(sn, seg.sn) > 0 {
			insert_idx = i + 1
			break
		}
	}

	if !repeat {
		dataCopy := zpool.GetBuffer2(len(newseg.data))
		copy(dataCopy, newseg.data)
		newseg.data = dataCopy

		if insert_idx == n+1 {
			kcp.rcv_buf = append(kcp.rcv_buf, newseg)
		} else {
			kcp.rcv_buf = append(kcp.rcv_buf, segment{})
			copy(kcp.rcv_buf[insert_idx+1:], kcp.rcv_buf[insert_idx:])
			kcp.rcv_buf[insert_idx] = newseg
		}
	}
	count := 0
	for k := range kcp.rcv_buf {
		seg := &kcp.rcv_buf[k]
		if seg.sn == kcp.rcv_nxt && len(kcp.rcv_queue)+count < int(kcp.rcv_wnd) {
			kcp.rcv_nxt++
			count++
		} else {
			break
		}
	}
	if count > 0 {
		kcp.rcv_queue = append(kcp.rcv_queue, kcp.rcv_buf[:count]...)
		kcp.rcv_buf = kcp.remove_front(kcp.rcv_buf, count)
	}
	return repeat
}
func (kcp *KCP) Input(data []byte, regular, ackNodelay bool) int {
	snd_una := kcp.snd_una
	if len(data) < IKCP_OVERHEAD {
		return -1
	}
	var latest uint32
	var flag int
	var inSegs uint64
	var windowSlides bool

	for {
		var ts, sn, length, una, conv uint32
		var wnd uint16
		var cmd, frg uint8
		if len(data) < int(IKCP_OVERHEAD) {
			break
		}
		data = ikcp_decode32u(data, &conv)
		if conv != kcp.conv {
			return -1
		}
		data = ikcp_decode8u(data, &cmd)
		data = ikcp_decode8u(data, &frg)
		data = ikcp_decode16u(data, &wnd)
		data = ikcp_decode32u(data, &ts)
		data = ikcp_decode32u(data, &sn)
		data = ikcp_decode32u(data, &una)
		data = ikcp_decode32u(data, &length)

		if len(data) < int(length) {
			return -2
		}
		if cmd != IKCP_CMD_PUSH && cmd != IKCP_CMD_ACK &&
			cmd != IKCP_CMD_WASK && cmd != IKCP_CMD_WINS {
			return -3
		}
		if regular {
			kcp.rmt_wnd = uint32(wnd)
		}
		if kcp.parse_una(una) > 0 {
			windowSlides = true
		}

		kcp.shrink_buf()

		if cmd == IKCP_CMD_ACK {
			kcp.parse_ack(sn)
			kcp.parse_fastack(sn, ts)
			flag |= 1
			latest = ts
		} else if cmd == IKCP_CMD_PUSH {
			repeat := true
			if _itimediff(sn, kcp.rcv_nxt+kcp.rcv_wnd) < 0 {
				kcp.ack_push(sn, ts)
				if _itimediff(sn, kcp.rcv_nxt) >= 0 {
					var seg segment
					seg.conv = conv
					seg.cmd = cmd
					seg.frg = frg
					seg.wnd = wnd
					seg.ts = ts
					seg.sn = sn
					seg.una = una
					seg.data = data[:length]
					repeat = kcp.parse_data(seg)
				}
			}
			if regular && repeat {
				atomic.AddUint64(&DefaultSnmp.RepeatSegs, 1)
			}
		} else if cmd == IKCP_CMD_WASK {
			kcp.probe |= IKCP_ASK_TELL
		} else if cmd == IKCP_CMD_WINS {

		} else {
			return -1
		}
		inSegs++
		data = data[length:]
	}
	atomic.AddUint64(&DefaultSnmp.InSegs, inSegs)

	if flag != 0 && regular {
		current := currentMs()
		if _itimediff(current, latest) >= 0 {
			kcp.update_ack(_itimediff(current, latest))
		}
	}
	if kcp.nocwnd == 0 {
		if _itimediff(kcp.snd_una, snd_una) > 0 {
			kcp
		}
	}
}

func (kcp *KCP) AAAAAAAAA() {

}
