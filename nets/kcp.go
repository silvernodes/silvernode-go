package nets

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/xtaci/kcp-go"
)

type KcpNetWorker struct {
}

func NewKcpNetWorker() *KcpNetWorker {
	k := new(KcpNetWorker)
	return k
}

func (k *KcpNetWorker) Listen(url string) error {
	url = strings.Trim(url, "udp://") // trim the ws header
	infos := strings.Split(url, "/")  // parse the sub path
	listener, err := kcp.Listen(infos[0])
	if err != nil {
		return err
	}
	defer listener.Close()
	boss := process.SpawnS()
	boss.Start(func() {
		conn, err := listener.Accept()
		if err != nil {
			k.onError(conn, err)
			boss.Terminate()
			return
		}
		worker := process.SpawnS()
		k.h_kcpSocket(conn, worker)
	}, nil)
	boss.Sync()
	return nil
}

func (k *KcpNetWorker) h_kcpSocket(conn net.Conn, worker process.Service) {
	buf := make([]byte, 4096, 4096)
	worker.Start(func() {
		n, err := conn.Read(buf[0:])
		if err != nil {
			k.onError(conn, err)
			worker.Terminate()
			return
		}
		if n > 0 {
			nodeId, exists := _connectManager.GetNodeIdByConn(conn)
			var temp []byte = make([]byte, 0, n)
			datas := bytes.NewBuffer(temp)
			datas.Write(buf[0:n])
			if exists {
				k.onMsg(conn, nodeId, datas.Bytes())
			} else {
				if err := k.dealHandShake(conn, worker, datas.Bytes()); err != nil {
					k.onError(conn, err)
				}
			}
			temp = nil // dispose the temp buffer
		} else {
			k.onError(conn, errutil.New("UDP设备未收到任何数据!!"))
		}
	}, func() {
		buf = nil
	})
}

func (k *KcpNetWorker) Connect(nodeId string, url string, origin string) error {
	theUrl := strings.Trim(url, "kcp://") // trim the header
	infos := strings.Split(theUrl, "/")   // parse the sub path
	conn, err := kcp.Dial(infos[0])
	if err != nil {
		return err
	}
	worker := process.SpawnS()
	if err := k.doHandShake(conn, worker, origin, url, nodeId); err != nil {
		worker.Terminate()
		return err
	}
	k.h_kcpSocket(conn, worker)
	return nil
}

func (k *KcpNetWorker) Send(conn net.Conn, msg []byte) error {
	defer func() {
		msg = nil // dispose the send buffer
	}()
	_, err := conn.Write(msg)
	return err
}

func (k *KcpNetWorker) onConn(conn net.Conn, worker process.Service, nodeId string, url string) {
	_, err := _connectManager.AddConnectInfo(nodeId, url, UDP, conn, worker, k)
	if err != nil {
		k.onError(conn, err)
	} else {
		_eventListener.OnConnect(nodeId)
	}
}

func (k *KcpNetWorker) onMsg(conn net.Conn, nodeId string, msg []byte) {
	if !_connectManager.CheckPingPong(nodeId, msg) {
		_eventListener.OnMessage(nodeId, msg)
	}
}

func (k *KcpNetWorker) onClose(nodeId string, conn net.Conn, reason error) {
	_eventListener.OnClose(nodeId, reason)
	_connectManager.RemoveConnectInfo(nodeId, conn)
	conn.Close()
}

func (k *KcpNetWorker) onError(conn net.Conn, err error) {
	if conn != nil {
		nodeId, exists := _connectManager.GetNodeIdByConn(conn)
		if exists {
			k.onClose(nodeId, conn, err) // close the conn with errors
		} else {
			conn.Close()
			_eventListener.OnError(err)
		}
	} else {
		_eventListener.OnError(err)
	}
}

func (k *KcpNetWorker) Close(nodeId string, conn net.Conn, err error) error {
	_eventListener.OnClose(nodeId, err)
	_connectManager.RemoveConnectInfo(nodeId, conn)
	return conn.Close()
}

func (k *KcpNetWorker) doHandShake(conn net.Conn, worker process.Service, origin string, url string, nodeId string) error {
	info := make(map[string]string)
	info["Header"] = "SILVERNODE/UDP"
	info["Origin"] = origin
	datas, err := json.Marshal(info)
	if err != nil {
		return err
	}
	if _, err2 := conn.Write(datas); err2 != nil {
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
		return errutil.New("UDP设备握手验证超时!!")
	}
	if buf[0] == 35 { // '#'
		strmsg := string(buf)
		if strmsg == "#hsuc" {
			k.onConn(conn, worker, nodeId, url)
			return nil
		}
	}
	return errutil.New("UDP设备收到非法的握手验证信息!!")
}

func (k *KcpNetWorker) dealHandShake(conn net.Conn, worker process.Service, msg []byte) error {
	var datas map[string]string
	if err := json.Unmarshal(msg, &datas); err != nil {
		return err
	}
	origin, exists := datas["Origin"]
	if !exists {
		return errutil.New("UDP握手验证信息丢失!")
	}
	nodeId, err := _eventListener.OnCheckNode(origin) // let the gonode to check if the url is legal
	if err != nil {
		return errutil.Extend("UDP设备收到非法的握手验证信息!!", err)
	}
	if _, err2 := conn.Write([]byte("#hsuc")); err2 != nil {
		return errutil.Extend("UDP设备握手验证信息回复失败", err)
	}
	k.onConn(conn, worker, nodeId, LOCAL)
	return nil
}
