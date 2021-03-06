package web

import (
	"github.com/BurntSushi/toml"
	"github.com/Yimismi/tools/tool"
	"github.com/gin-gonic/gin"
	"github.com/ngaut/log"
	"io"
	"net/http"
	"os"
	"path"
)

var webTool map[string]tool.WebTool

type ServerConfig struct {
	Port         string      `toml:"port"`
	LogRoot      string      `toml:"log_root"`
	ExternalTool []tool.Tool `toml:"external_tool"`
}

func init() {
	webTool = make(map[string]tool.WebTool)
}
func RegisterTool(t tool.WebTool) {
	webTool[t.GetName()] = t
}

func Run(confPath string) {
	conf := loadConfig(confPath)
	setLog(conf)
	r := gin.Default()
	loadStaticFs(r, conf)
	loadWebTool(r)
	r.Run(conf.Port)
}
func setLog(conf *ServerConfig) {
	f, err := os.Create(path.Join(conf.LogRoot, "gin.log"))
	if err != nil {
		panic(err)
	}
	gin.DefaultWriter = io.MultiWriter(f)
}
func loadConfig(confPath string) *ServerConfig {
	filePath := path.Join(confPath, "server.conf")
	sc := new(ServerConfig)
	_, err := toml.DecodeFile(filePath, sc)
	if err != nil {
		panic(err)
		return nil
	} else {
		log.Infof("load config succeed:\n%+v\n", sc)
	}
	return sc
}
func loadStaticFs(r *gin.Engine, conf *ServerConfig) {
	r.LoadHTMLGlob("./web/view/*.tmpl")
	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Tools":   webTool,
			"ExTools": conf.ExternalTool,
		})
	})
	r.GET("/", func(c *gin.Context) {
		c.Redirect(303, "/index")
	})
	r.Static("/js", "./web/js/")
	r.Static("/css", "./web/css/")
	r.Static("/tool", "./web/view/static/tool")
}

func loadWebTool(r *gin.Engine) {
	for name, tool := range webTool {
		r.POST(tool.GetUrl(), func(context *gin.Context) {
			webProcessor(name, context)
		})
		r.GET(path.Join("/args", tool.GetName()), func(context *gin.Context) {
			context.JSON(http.StatusOK, tool.GetArgsDesc())
		})
	}
}

func webProcessor(name string, c *gin.Context) {
	tool, ok := webTool[name]
	if !ok {
		c.JSON(200, map[string]string{"error": tool.GetName() + "is not found"})
		return
	}
	tool.Run(c)
}
