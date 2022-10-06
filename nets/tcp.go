package nets

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/buffutil"
	"github.com/silvernodes/silvernode-go/utils/errutil"
)

const (
	PCK_MIN_SIZE int   = 6          // |--- header 4bytes ---|--- length 2 bytes ---|--- other datas --- ....
	PCK_HEADER   int32 = 0x2123676f // !#go
)

type TcpNetWorker struct {
}

func NewTcpNetWorker() *TcpNetWorker {
	t := new(TcpNetWorker)
	return t
}

func (t *TcpNetWorker) Listen(url string) error {
	url = strings.Trim(url, "tcp://") // trim the ws header
	infos := strings.Split(url, "/")  // parse the sub path
	tcpAddr, err := net.ResolveTCPAddr("tcp", infos[0])
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	defer listener.Close()
	boss := process.SpawnS()
	boss.Start(func() {
		conn, err := listener.Accept()
		if err != nil {
			t.onError(conn, err)
			boss.Terminate()
			return
		}
		worker := process.SpawnS()
		t.h_tcpSocket(conn, worker)
	}, nil)
	boss.Sync()
	return nil
}

func (t *TcpNetWorker) h_tcpSocket(conn net.Conn, worker process.Service) {
	buf := make([]byte, 8192, 8192)
	rcvbuf := NewTcpBuffer(buf)
	worker.Start(func() {
		n, err := conn.Read(rcvbuf.Buffer())
		if err != nil {
			t.onError(conn, err)
			worker.Terminate()
			return
		}
		if n > 0 {
			rcvbuf.AddDataLen(n)
			for rcvbuf.Count() > PCK_MIN_SIZE {
				parser := buffutil.NewParser(rcvbuf.Slice(), 0)
				head := parser.ReadInt()
				if parser.Error() != nil || head != PCK_HEADER {
					rcvbuf.Clear()
					break
				}
				l := parser.ReadShort()
				length := int(l)
				if parser.Error() != nil {
					rcvbuf.Clear()
					break
				} else if length > rcvbuf.Count() {
					break
				}
				nodeId, exists := _connectManager.GetNodeIdByConn(conn)
				src := rcvbuf.Slice()[PCK_MIN_SIZE : PCK_MIN_SIZE+length]
				var temp []byte = make([]byte, 0, length)
				datas := bytes.NewBuffer(temp)
				datas.Write(src)
				rcvbuf.DeleteData(length + PCK_MIN_SIZE)
				if exists {
					t.onMsg(conn, nodeId, datas.Bytes())
				} else {
					if err := t.dealHandShake(conn, worker, string(datas.Bytes())); err != nil {
						t.onError(conn, err)
						return
					}
				}
				temp = nil // dispose the temp buffer
			}
			rcvbuf.Reset()
		} else {
			t.onError(conn, errutil.New("TCP设备未收到任何数据!!"))
		}
	}, func() {
		rcvbuf.Dispose()
	})

	// or you can use the bufio.Scanner like t
	/*
		scanner := bufio.NewScanner(conn)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if len(data) > PCK_MIN_SIZE {
				if !atEOF {
					parser := binbuf.BuildParser(data, 0)
					header, err := parser.Int()
					if err == nil && header == PCK_HEADER {
						length, err2 := parser.Short()
						if err2 == nil {
							needlen := int(length) + PCK_MIN_SIZE
							if needlen <= len(data) { // parser a whole package
								return needlen, data[:needlen], nil
							}
						}
					}
				}
			}
			return
		})
		for scanner.Scan() {
			err := scanner.Err()
			if err != nil {
				t.onError(conn, err)
				break
			}
			nodeId, exists := nets.GetInfonodeIdByConn(conn)
			buf := scanner.Bytes()
			var temp []byte = make([]byte, 0, len(buf)-PCK_MIN_SIZE)
			datas := bytes.NewBuffer(temp)
			datas.Write(buf[PCK_MIN_SIZE:])
			if exists {
				t.onMsg(conn, nodeId, datas.Bytes())
			} else {
				if err := t.dealHandShake(conn, string(datas.Bytes())); err != nil {
					t.onError(conn, err)
				}
			}
		}
	*/
}

func (t *TcpNetWorker) Connect(nodeId string, url string, origin string) error {
	theUrl := strings.Trim(url, "tcp://") // trim the ws header
	infos := strings.Split(theUrl, "/")   // parse the sub path
	tcpAddr, err := net.ResolveTCPAddr("tcp", infos[0])
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	worker := process.SpawnS()
	if err := t.doHandShake(conn, worker, origin, url, nodeId); err != nil {
		return err
	}
	t.h_tcpSocket(conn, worker)
	return nil
}

func (t *TcpNetWorker) Send(conn net.Conn, msg []byte) error {
	datalen := len(msg)
	buf := buffutil.NewBuffer(datalen + PCK_MIN_SIZE)
	defer func() {
		msg = nil // dispose the send buffer
		buf.Dispose()
		buf = nil
	}()
	bytes, err := buf.WriteInt(PCK_HEADER).WriteShort(int16(datalen)).WriteBytes(msg).Flush()
	if err != nil {
		return err
	}
	_, err2 := conn.Write(bytes)
	return err2
}

func (t *TcpNetWorker) onConn(conn net.Conn, worker process.Service, nodeId string, url string) {
	// record the set from nodeId to conn
	_, err := _connectManager.AddConnectInfo(nodeId, url, TCP, conn, worker, t)
	if err != nil {
		t.onError(conn, err)
	} else {
		_eventListener.OnConnect(nodeId)
	}
}

func (t *TcpNetWorker) onMsg(conn net.Conn, nodeId string, msg []byte) {
	if !_connectManager.CheckPingPong(nodeId, msg) {
		_eventListener.OnMessage(nodeId, msg)
	}
}

func (t *TcpNetWorker) onClose(nodeId string, conn net.Conn, reason error) {
	_eventListener.OnClose(nodeId, reason)
	_connectManager.RemoveConnectInfo(nodeId, conn) // remove the closed conn from local record
	conn.Close()
}

func (t *TcpNetWorker) onError(conn net.Conn, err error) {
	if conn != nil {
		nodeId, exists := _connectManager.GetNodeIdByConn(conn)
		if exists {
			t.onClose(nodeId, conn, err) // close the conn with errors
		} else {
			conn.Close()
			if !errutil.IsEOF(err) {
				_eventListener.OnError(err)
			}
		}
	} else {
		_eventListener.OnError(err)
	}
}

func (t *TcpNetWorker) Close(nodeId string, conn net.Conn, err error) error {
	_eventListener.OnClose(nodeId, err)
	_connectManager.RemoveConnectInfo(nodeId, conn)
	return conn.Close()
}

func (t *TcpNetWorker) doHandShake(conn net.Conn, worker process.Service, origin string, url string, nodeId string) error {
	info := make(map[string]string)
	info["Header"] = "SILVERNODE/TCP"
	info["Origin"] = origin
	datas, err := json.Marshal(info)
	if err != nil {
		return err
	}
	if err2 := t.Send(conn, datas); err2 != nil {
		return err2
	}

	buf := make([]byte, 5, 5) // the rev buf
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * 6)); err != nil {
		return err
	}
	n, err := conn.Read(buf[0:])
	if err != nil {
		return err
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}
	if n < 0 {
		return errutil.New("TCP设备握手验证超时!!")
	}
	if buf[0] == 35 { // '#'
		strmsg := string(buf)
		if strmsg == "#hsuc" {
			t.onConn(conn, worker, nodeId, url)
			return nil
		}
	}
	return errutil.New("TCP设备收到非法的握手验证信息!!")
}

func (t *TcpNetWorker) dealHandShake(conn net.Conn, worker process.Service, info string) error {
	var datas map[string]string
	if err := json.Unmarshal([]byte(info), &datas); err != nil {
		return err
	}
	origin, exists := datas["Origin"]
	if !exists {
		return errutil.New("TCP握手验证信息丢失!")
	}
	nodeId, err := _eventListener.OnCheckNode(origin) // let the gonode to check if the url is legal
	if err != nil {
		return errutil.Extend("TCP设备收到非法的握手验证信息!!", err)
	}
	if _, err2 := conn.Write([]byte("#hsuc")); err2 != nil {
		return errutil.Extend("TCP设备握手验证信息回复失败", err)
	}
	t.onConn(conn, worker, nodeId, LOCAL)
	return nil
}

type TcpBuffer struct {
	_count  int
	_offset int
	_buffer []byte
	_len    int
}

func NewTcpBuffer(buf []byte) *TcpBuffer {
	t := new(TcpBuffer)
	t._buffer = buf
	t._offset = 0
	t._count = 0
	t._len = len(t._buffer)
	return t
}

func (t *TcpBuffer) Clear() {
	t._offset = 0
	t._count = 0
	for i := 0; i < t._len; i++ {
		t._buffer[i] = byte(0)
	}
}

func (t *TcpBuffer) Reset() {
	copy(t._buffer, t.Slice())
	t._offset = 0
}

func (t *TcpBuffer) Buffer() []byte {
	return t._buffer[t._offset:]
}

func (t *TcpBuffer) Slice() []byte {
	return t._buffer[t._offset : t._offset+t._count]
}

func (t *TcpBuffer) Count() int {
	return t._count
}

func (t *TcpBuffer) Offset() int {
	return t._offset
}

func (t *TcpBuffer) Capcity() int {
	return t._len - t._offset
}

func (t *TcpBuffer) AddDataLen(count int) {
	t._count += count
}

func (t *TcpBuffer) DeleteData(count int) {
	if t._count >= count {
		t._offset += count
		t._count -= count
	}
}

func (t *TcpBuffer) Dispose() {
	t._buffer = nil
	t._offset = 0
	t._count = 0
	t._len = 0
}
