package http

import (
	"fmt"
	_http "net/http"
	"reflect"
	"strings"

	_proc "github.com/silvernodes/silvernode-go/peers/proc"
)

type Server struct {
	codec   Codec
	httpMux *_http.ServeMux
	_server *_http.Server
}

func NewServer() *Server {
	s := new(Server)
	s.codec = NewPostJsonCodec()
	s.httpMux = _http.NewServeMux()
	return s
}

func (s *Server) SetCodec(codec Codec) {
	s.codec = codec
}

func (s *Server) Route(path string, proc interface{}) {
	typ := reflect.TypeOf(proc)
	val := reflect.ValueOf(proc)
	meta, ok := _proc.CheckHasMeta(typ, val)
	_proc.WalkSuitableMethods(proc, func(method reflect.Method, argType, replyType reflect.Type) {
		url := path
		if !strings.HasPrefix(url, "/") {
			url += "/"
		}
		if method.Name != "Index" {
			url += "/" + method.Name
		}
		s.httpMux.HandleFunc(url, func(w _http.ResponseWriter, r *_http.Request) {
			defer r.Body.Close()
			if ok {
				args, reply, err := meta.CreateBeans(method.Name, "", r)
				if err != nil {
					w.WriteHeader(_http.StatusInternalServerError)
					fmt.Fprintln(w, err.Error())
					return
				}
				if err := s.codec.Decode(r, args); err != nil {
					// 数据解析出错
					w.WriteHeader(_http.StatusInternalServerError)
					fmt.Fprintln(w, "数据解析出错:"+err.Error())
					return
				}
				if err := meta.ProcessFlow(method.Name, proc, args, reply); err != nil {
					w.WriteHeader(_http.StatusInternalServerError)
					fmt.Fprintln(w, string(err.Error()))
					return
				}
				retdata, err := s.codec.Encode(reply)
				if err != nil {
					// 解析错误
					w.WriteHeader(_http.StatusInternalServerError)
					fmt.Fprintln(w, "应答数据序列化出错")
					return
				}
				w.Write(retdata)
			} else {
				argv := reflect.New(argType.Elem())
				if err := s.codec.Decode(r, argv.Interface()); err != nil {
					// 数据解析出错
					w.WriteHeader(_http.StatusInternalServerError)
					fmt.Fprintln(w, "数据解析出错:"+err.Error())
					return
				}
				replyv := reflect.New(replyType.Elem())
				switch replyType.Elem().Kind() {
				case reflect.Map:
					replyv.Elem().Set(reflect.MakeMap(replyType.Elem()))
				case reflect.Slice:
					replyv.Elem().Set(reflect.MakeSlice(replyType.Elem(), 0, 0))
				}
				procv := reflect.ValueOf(proc)
				returnValues := method.Func.Call([]reflect.Value{procv, argv, replyv})
				errInter := returnValues[0].Interface()
				var reterr error = nil
				if errInter != nil {
					reterr = errInter.(error)
					// 返回错误
					w.WriteHeader(_http.StatusInternalServerError)
					fmt.Fprintln(w, reterr.Error())
					return
				}
				retdata, err := s.codec.Encode(replyv)
				if err != nil {
					// 解析错误
					w.WriteHeader(_http.StatusInternalServerError)
					fmt.Fprintln(w, "应答数据序列化出错")
					return
				}
				w.Write(retdata)
			}
		})
		_proc.RecordMeta(proc)
		txt := "路由[" + url + "]映射完毕."
		if ok {
			txt += "(!)"
		}
		fmt.Println(txt)
	})
}

func (s *Server) Listen(addr string) error {
	s._server = &_http.Server{
		Addr:    addr,
		Handler: s.httpMux,
	}
	return s._server.ListenAndServe()
}

func (s *Server) Close() error {
	if s._server != nil {
		return s._server.Close()
	}
	return nil
}
