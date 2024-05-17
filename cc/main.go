package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/titanous/json5"
	"gogs.mikescher.com/BlackForestBytes/goext/exerr"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	json "gogs.mikescher.com/BlackForestBytes/goext/gojson"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"gogs.mikescher.com/BlackForestBytes/goext/rfctime"
	"gogs.mikescher.com/BlackForestBytes/goext/timeext"
	"html/template"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"
)

//go:embed index.html
var indexHTML string

//go:embed history.html
var historyHTML string

var commandHistory = make([]Command, 0)
var commands = make([]Command, 0)
var gil = sync.Mutex{}

var dataDir string
var lastSaveData string

func main() {
	cw := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05 Z07:00",
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	multi := zerolog.MultiLevelWriter(cw)
	logger := zerolog.New(multi).With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = logger

	gin.SetMode(gin.DebugMode)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	exerr.Init(exerr.ErrorPackageConfigInit{})

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	dataDir = filepath.Dir(ex)

	if v, ok := os.LookupEnv("CC_DATADIR"); ok {
		dataDir = v
	}

	err = os.MkdirAll(dataDir, 0777)
	if err != nil {
		panic(err)
	}

	loadCommands(true)

	engine := ginext.NewEngine(ginext.Options{})

	engine.Routes().GET("/").Handle(indexPage)
	engine.Routes().GET("/history").Handle(historyPage)

	engine.Routes().POST("/cc").Handle(popCommands)
	engine.Routes().GET("/cc").Handle(peekCommands)
	engine.Routes().PUT("/cc").Handle(addCommands)

	engine.Routes().POST("/templates/clubscale-tickets").Handle(addTemplateClubscaleTickets)

	routerErr, router := engine.ListenAndServeHTTP("0.0.0.0:8000", func(port string) {})

	sigstop := make(chan os.Signal, 1)
	signal.Notify(sigstop, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case err := <-routerErr:
		log.Error().Err(err).Msg("router finished with error")
		panic(err)
	case <-sigstop:
		log.Info().Msg("Received SIGTERM/SIGKILL - Stopping all routers+app")
		err := router.Shutdown(context.Background())
		if err != nil {
			log.Err(err).Msg("failed to stop reverse-router")
		}
	}

	log.Info().Msg("Server finished.")
}

func updateCommands(lock bool) {

	if lock {
		gil.Lock()
		defer gil.Unlock()
	}

	ncmd := make([]Command, 0, len(commands))

	for _, cmd := range commands {

		if time.Now().Unix()-cmd.Date.Unix() > 60 { // remove commands older than 60s

			cmd.Status = "SKIPPED"
			commandHistory = append(commandHistory, cmd)

			log.Info().Msgf("[SKIP]  %s", cmd.String())
		} else {
			ncmd = append(ncmd, cmd)
		}

	}

	commands = ncmd

	saveCommands(false)
}

func saveCommands(lock bool) {

	if lock {
		gil.Lock()
		defer gil.Unlock()
	}

	str := ""

	for _, cmd := range commandHistory {

		b, err := json.Marshal(cmd.Serialize(false))
		if err != nil {
			log.Err(err).Msg("failed to marshal command")
			return
		}

		str += string(b) + "\n"

	}

	for _, cmd := range commands {

		b, err := json.Marshal(cmd.Serialize(false))
		if err != nil {
			log.Err(err).Msg("failed to marshal command")
			return
		}

		str += string(b) + "\n"

	}

	fp := filepath.Join(dataDir, "commands.jsonl")

	if str == lastSaveData {
		return
	}

	err := os.WriteFile(fp, []byte(str), 0777)
	if err != nil {
		log.Err(err).Msg("failed to write cmds to " + fp)
		return
	}

	lastSaveData = str

	fmt.Printf("[SAV] Written [%d+%d] commands (%d bytes) to '%s'\n", len(commands), len(commandHistory), len([]byte(str)), fp)
}

func loadCommands(lock bool) {

	if lock {
		gil.Lock()
		defer gil.Unlock()
	}

	fp := filepath.Join(dataDir, "commands.jsonl")

	if !langext.FileExists(fp) {
		fmt.Printf("Loaded 0 commands from '%s' (file does not exist)\n", fp)
		commands = make([]Command, 0)
		return
	}

	ncmd := make([]Command, 0)
	hcmd := make([]Command, 0)

	bin, err := os.ReadFile(fp)
	if err != nil {
		panic("ERR: failed to read commands from " + fp + ":\n" + err.Error())
	}

	for _, line := range strings.Split(string(bin), "\n") {

		if strings.TrimSpace(line) == "" {
			continue
		}

		cmd, err := DeserializeCommand([]byte(line))
		if err != nil {
			panic("ERR: failed to read commands from " + fp + ":\n" + err.Error() + "\n" + line)
		}

		if cmd.Status == "PENDING" {
			ncmd = append(ncmd, cmd)
		} else {
			hcmd = append(hcmd, cmd)
		}

	}

	commands = ncmd
	commandHistory = hcmd
	lastSaveData = string(bin)

	fmt.Printf("Loaded [%d+%d] commands from '%s'\n", len(ncmd), len(hcmd), fp)
}

func addCommands(pctx ginext.PreContext) ginext.HTTPResponse {
	var rb []byte
	_, gctx, errResp := pctx.RawBody(&rb).Start()
	if errResp != nil {
		return *errResp
	}

	return addNewCommands(rb, gctx)
}

func addNewCommands(rb []byte, gctx *gin.Context) ginext.HTTPResponse {
	t0 := time.Now()

	type body []Command

	var b body
	if strings.HasPrefix(strings.TrimSpace(string(rb)), "{") {
		var sb Command
		err := json5.Unmarshal(rb, &sb)
		if err != nil {
			return ginext.Error(err)
		}
		b = body{sb}
	} else {
		err := json5.Unmarshal(rb, &b)
		if err != nil {
			return ginext.Error(err)
		}
	}

	log.Debug().Msgf("[REQ]  <add> from %s", gctx.ClientIP())

	for _, cmd := range b {
		if msg, ok := cmd.Valid(); !ok {
			return ginext.Text(400, msg)
		}
	}

	gil.Lock()
	defer gil.Unlock()

	updateCommands(false)

	for _, cmd := range b {
		cmd.ID = langext.MustHexUUID()
		cmd.Date = t0
		cmd.Status = "PENDING"
		cmd.Executed = nil

		commands = append(commands, cmd)

		log.Info().Msgf("[ADD]  %s", cmd.String())
	}

	saveCommands(false)

	return ginext.Text(200, fmt.Sprintf("OK - added %d commands, %d commands enqueued", len(b), len(commands)))
}

func popCommands(pctx ginext.PreContext) ginext.HTTPResponse {
	_, gctx, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}

	log.Debug().Msgf("[REQ]  <pop> from %s", gctx.ClientIP())

	gil.Lock()
	defer gil.Unlock()

	updateCommands(false)

	resp := ""
	for _, cmd := range commands {

		cmd.Status = "EXECUTED"
		cmd.Executed = langext.Ptr(time.Now())
		commandHistory = append(commandHistory, cmd)

		resp += cmd.String() + "\n"
		log.Info().Msgf("[POP]  %s", cmd.String())

	}

	commands = make([]Command, 0)

	return ginext.Text(200, resp)
}

func peekCommands(pctx ginext.PreContext) ginext.HTTPResponse {
	_, gctx, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}

	log.Debug().Msgf("[REQ]  <peek> from %s", gctx.ClientIP())

	gil.Lock()
	defer gil.Unlock()

	updateCommands(false)

	resp := ""
	for _, cmd := range commands {
		resp += cmd.String() + "\n"
		log.Info().Msgf("[PEEK] %s", cmd.String())
	}

	return ginext.Text(200, resp)
}

func indexPage(pctx ginext.PreContext) ginext.HTTPResponse {
	_, _, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}

	templ, err := template.New("index").Funcs(tepmlateFuncs()).Parse(indexHTML)
	if err != nil {
		return ginext.Error(err)
	}

	updateCommands(true)

	commandsCopy := func() []Command {
		gil.Lock()
		defer gil.Unlock()
		r := langext.ArrConcat(langext.ArrCopy(commandHistory), langext.ArrCopy(commands))
		langext.ReverseArray(r)
		return r
	}()

	data := gin.H{
		"Commands": commandsCopy,
	}

	buffer := bytes.Buffer{}
	err = templ.Execute(&buffer, data)
	if err != nil {
		return ginext.Error(err)
	}

	return ginext.Data(200, "text/html", buffer.Bytes())
}

func tepmlateFuncs() template.FuncMap {
	return template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) }, //nolint:gosec
		"json": func(obj any) string {
			v, err := json.Marshal(obj)
			if err != nil {
				panic(err)
			}
			return string(v)
		},
		"json_indent": func(obj any) string {
			v, err := json.MarshalIndent(obj, "", "  ")
			if err != nil {
				panic(err)
			}
			return string(v)
		},
		"deref": func(vInput any) any {
			val := reflect.ValueOf(vInput)
			if val.Kind() == reflect.Ptr {
				return val.Elem().Interface()
			}
			return "NO"
		},
		"date": func(v rfctime.AnyTime) string {
			if r, ok := v.(time.Time); ok {
				return r.In(timeext.TimezoneBerlin).Format("2006-01-02 15:04:05")
			}
			if r, ok := v.(rfctime.RFCTime); ok {
				return r.Time().In(timeext.TimezoneBerlin).Format("2006-01-02 15:04:05")
			}
			return time.Unix(0, v.UnixNano()).In(timeext.TimezoneBerlin).Format("2006-01-02 15:04:05")
		},
	}
}

func historyPage(pctx ginext.PreContext) ginext.HTTPResponse {
	_, _, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}

	templ, err := template.New("history").Funcs(tepmlateFuncs()).Parse(historyHTML)
	if err != nil {
		return ginext.Error(err)
	}

	updateCommands(true)

	commandsCopy := func() []Command {
		gil.Lock()
		defer gil.Unlock()
		r := langext.ArrConcat(langext.ArrCopy(commandHistory), langext.ArrCopy(commands))
		langext.ReverseArray(r)
		return r
	}()

	data := gin.H{
		"Commands": commandsCopy,
	}

	buffer := bytes.Buffer{}
	err = templ.Execute(&buffer, data)
	if err != nil {
		return ginext.Error(err)
	}

	return ginext.Data(200, "text/html", buffer.Bytes())
}

func addTemplateClubscaleTickets(pctx ginext.PreContext) ginext.HTTPResponse {
	_, gctx, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}

	cmds := `
[


    {    "action":"LAMP",            "delay": 0,                 "duration": 8000,       },

    {
        "action":"BUZZER_PWM_NOTES",     
        "delay": 0,                      
        "noteLength": 100,               
        "notes": [                       
            1500, 2000, 2500, 3000,
            1500, 2000, 2500, 3000,
            1500, 2000, 2500, 3000,
            2000, 2000, 2000, 2000,
        ],
    },

    {    "action":"BUZZER_1",        "delay": 2000,   "duration": 200,        },
    {    "action":"BUZZER_1",        "delay": 2400,   "duration": 200,        },
    {    "action":"BUZZER_1",        "delay": 2800,   "duration": 200,        },

    {    "action":"BUZZER_1",        "delay": 4000,   "duration": 200,        },
    {    "action":"BUZZER_1",        "delay": 4400,   "duration": 200,        },
    {    "action":"BUZZER_1",        "delay": 4800,   "duration": 200,        },

]
`

	return addNewCommands([]byte(cmds), gctx)
}
