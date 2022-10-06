package proc

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/fileutil"
	"github.com/silvernodes/silvernode-go/utils/stringutil"
)

func getPkgOnDisk(pkg string) string {
	curDir := fileutil.CurrentDir()
	if pkg == "main" {
		return curDir
	}
	tmps := strings.Split(curDir, "/")
	lst := tmps[len(tmps)-2]
	if !strings.Contains(pkg, lst) {
		return curDir
	}
	tmps2 := strings.Split(pkg, lst+"/")
	tmp := curDir + tmps2[len(tmps2)-1]
	tmp = strings.ReplaceAll(tmp, "\\", "/")
	return tmp
}

func getPkgName(pkg string) string {
	infos := strings.Split(pkg, "/")
	return infos[len(infos)-1]
}

func removePeerMeta(p *_MetaInfo) {
	path := p.typ.Elem().PkgPath()
	disk := getPkgOnDisk(path)
	name := p.typ.Elem().Name()
	file := stringutil.Snake(name) + "_meta.go"

	if fileutil.Exists(disk + file) {
		if err := fileutil.DeleteFile(disk + file); err != nil {
			fmt.Println("[" + file + "]被移除.")
		}
	}
}

func buildPeerMeta(p *_MetaInfo) error {
	path := p.typ.Elem().PkgPath()
	disk := getPkgOnDisk(path)
	name := p.typ.Elem().Name()
	file := stringutil.Snake(name) + "_meta.go"
	metaName := name + "Meta"

	pkgName := getPkgName(path)
	txt := "package " + pkgName + "\r\n\r\n"
	txt += "import (\r\n"
	txt += "\t\"errors\"\r\n"
	txt += "\t\"github.com/silvernodes/silvernode-go/peers/proc\"\r\n"
	txt += "%s"
	txt += ")\r\n\r\n"
	txt += "type " + metaName + " struct {\r\n}\r\n"
	txt += "\r\n"
	txt += "func (m *" + metaName + ") CreateBeans(method string, from string, ctx interface{}) (interface{}, interface{}, error) {\r\n"
	txt += "\tswitch method {\r\n"
	txt += "%s"
	txt += "\t}\r\n"
	txt += "\treturn nil, nil, errors.New(\"不支持的方法装配:\" + method)"
	txt += "\r\n}\r\n\r\n"
	txt += "func (m *" + metaName + ") ProcessFlow(method string, proc interface{}, args interface{}, reply interface{}) error {\r\n"
	txt += "\tswitch method {\r\n"
	txt += "%s"
	txt += "\t}\r\n"
	txt += "\treturn errors.New(\"不支持的方法调用:\" + method)"
	txt += "\r\n}\r\n\r\n"
	txt += "func (p *" + name + ") GetMeta() proc.ProcMeta {\r\n"
	txt += "\treturn new(" + metaName + ")\r\n"
	txt += "}\r\n"

	pkgs, bean, invoke := flushFuncBody(p)
	txt = fmt.Sprintf(txt, pkgs, bean, invoke)

	if err := fileutil.SaveFile(disk+"/"+file, txt); err != nil {
		return errutil.Extend("发布meta文件出错", err)
	}

	fmt.Println("[" + file + "]生成完毕.")
	return nil
}

func flushFuncBody(p *_MetaInfo) (string, string, string) {
	pkgList := make(map[string]byte)
	pkgs := ""
	bean := ""
	invoke := ""
	typ := p.typ.Elem().Name()
	pkg := p.typ.Elem().PkgPath()
	for name, mtype := range p.methods {
		argsType := mtype.ArgType.Elem().Name()
		argsPkg := mtype.ArgType.Elem().PkgPath()
		replyType := mtype.ReplyType.Elem().Name()
		replyPkg := mtype.ReplyType.Elem().PkgPath()

		if argsPkg != pkg {
			pkgList[argsPkg] = 1
			pkgName := getPkgName(argsPkg)
			argsType = pkgName + "." + argsType
		}
		if replyPkg != pkg {
			pkgList[replyPkg] = 1
			pkgName := getPkgName(replyPkg)
			replyType = pkgName + "." + replyType
		}

		bean += "\tcase \"" + name + "\":\r\n"
		bean += "\t\targs := new(" + argsType + ")\r\n"
		if txt := autoWiredArgsFields(mtype.ArgType.Elem()); txt != "" {
			bean += txt
		}
		bean += "\t\treply := new(" + replyType + ")\r\n"
		bean += "\t\treturn args, reply, nil\r\n"

		invoke += "\tcase \"" + name + "\":\r\n"
		invoke += "\t\treturn proc.(*" + typ + ")." + name + "(args.(*" + argsType + "), reply.(*" + replyType + "))\r\n"
	}
	for pkg, _ := range pkgList {
		pkgs += "\t\"" + pkg + "\"\r\n"
	}
	return pkgs, bean, invoke
}

func autoWiredArgsFields(argt reflect.Type) string {
	numfield := argt.NumField()
	if numfield <= 0 {
		return ""
	}
	txt := ""
	for index := 0; index < numfield; index++ {
		field := argt.Field(index)
		auto := field.Tag.Get("auto")
		if auto == "node" {
			if field.Type.Kind() == reflect.String {
				txt += "\t\targs." + field.Name + " = from\r\n"
			}
		} else if auto == "ctx" {
			typeKind := field.Type.Kind()
			if typeKind == reflect.Ptr || typeKind == reflect.Interface {
				txt += "\t\targs." + field.Name + " = ctx\r\n"
			}
		}
	}
	return txt
}
