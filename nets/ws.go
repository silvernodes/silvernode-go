package nets

import (
	"net"
	"net/http"
	"strings"

	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"golang.org/x/net/websocket"
)

type WSNetWorker struct {
}

func NewWSNetWorker() *WSNetWorker {
	w := new(WSNetWorker)
	return w
}

func (w *WSNetWorker) Listen(url string) error {
	url = strings.Trim(url, "ws://") // trim the ws header
	infos := strings.Split(url, "/") // parse the sub path
	wsMux := http.NewServeMux()
	wsMux.Handle("/"+infos[1], websocket.Handler(w.h_webSocket))
	err := http.ListenAndServe(infos[0], wsMux)
	return err
}

func (w *WSNetWorker) h_webSocket(conn *websocket.Conn) {
	remote := conn.RemoteAddr().String()
	nodeId, err := _eventListener.OnCheckNode(remote) // let the gonode to check if the url is legal
	if err == nil {
		worker := process.SpawnS()
		w.onConn(conn, worker, nodeId, LOCAL)
		var msg []byte
		worker.Start(func() {
			err := websocket.Message.Receive(conn, &msg)
			if err != nil {
				w.onError(conn, err)
				worker.Terminate()
				return
			}
			w.onMsg(conn, msg)
		}, nil)
		worker.Sync()
	} else {
		w.onError(conn, err)
	}
}

func (w *WSNetWorker) Connect(nodeId string, url string, origin string) error {
	conn, err := websocket.Dial(url, "tcp", origin)
	if err == nil {
		worker := process.SpawnS()
		w.onConn(conn, worker, nodeId, url)
		var msg []byte
		worker.Start(func() {
			err := websocket.Message.Receive(conn, &msg)
			if err != nil {
				w.onError(conn, err)
				worker.Terminate()
				return
			}
			w.onMsg(conn, msg)
		}, nil)
	}
	return err
}

func (w *WSNetWorker) Send(conn net.Conn, msg []byte) error {
	defer func() {
		msg = nil // dispose the send buffer
	}()
	err := websocket.Message.Send(conn.(*websocket.Conn), msg)
	return err
}

func (w *WSNetWorker) SendText(nodeId string, str string) error {
	info, exist := _connectManager.GetConnectInfo(nodeId)
	if !exist {
		return errutil.New("未能找到对应的链路信息:" + nodeId)
	}
	err := websocket.Message.Send(info.conn.(*websocket.Conn), str)
	return err
}

func (w *WSNetWorker) onConn(conn *websocket.Conn, worker process.Service, nodeId string, url string) {
	// record the set from nodeId to conn
	_, err := _connectManager.AddConnectInfo(nodeId, url, WS, conn, worker, w)
	if err != nil {
		w.onError(conn, err)
	} else {
		_eventListener.OnConnect(nodeId)
	}
}

func (w *WSNetWorker) onMsg(conn *websocket.Conn, msg []byte) {
	nodeId, exists := _connectManager.GetNodeIdByConn(conn)
	if exists {
		if msg[0] == 35 && len(msg) == 5 {
			strmsg := string(msg)
			if strmsg == "#ping" {
				go w.Send(conn, []byte("#pong"))
				return
			}
		}
		_eventListener.OnMessage(nodeId, msg)
	}
}

func (w *WSNetWorker) onClose(nodeId string, conn *websocket.Conn, reason error) {
	_eventListener.OnClose(nodeId, reason)
	_connectManager.RemoveConnectInfo(nodeId, conn) // remove the closed conn from local record
	conn.Close()
}

func (w *WSNetWorker) onError(conn *websocket.Conn, err error) {
	if conn != nil {
		nodeId, exists := _connectManager.GetNodeIdByConn(conn)
		if exists {
			w.onClose(nodeId, conn, err) // close the conn with errors
		} else {
			conn.Close()
			_eventListener.OnError(err)
		}
	} else {
		_eventListener.OnError(err)
	}
}

func (w *WSNetWorker) Close(nodeId string, conn net.Conn, err error) error {
	_eventListener.OnClose(nodeId, err)
	_connectManager.RemoveConnectInfo(nodeId, conn) // remove the closed conn from local record
	return conn.Close()
}
