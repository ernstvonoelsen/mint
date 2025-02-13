package app

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"

	"github.com/mintoolkit/mint/pkg/consts"
	"github.com/mintoolkit/mint/pkg/util/fsutil"
	v "github.com/mintoolkit/mint/pkg/version"
)

const (
	ofJSON         = "json"
	ofText         = "text"
	ofSubscription = "subscription"
)

type ExecutionContext struct {
	Out             *Output
	Args            []string
	cleanupHandlers []func()
}

func (ref *ExecutionContext) Exit(exitCode int) {
	ref.doCleanup()
	ref.exit(exitCode)
}

func (ref *ExecutionContext) AddCleanupHandler(handler func()) {
	if handler != nil {
		ref.cleanupHandlers = append(ref.cleanupHandlers, handler)
	}
}

func (ref *ExecutionContext) doCleanup() {
	if len(ref.cleanupHandlers) == 0 {
		return
	}

	//call cleanup handlers in reverse order
	for i := len(ref.cleanupHandlers) - 1; i >= 0; i-- {
		cleanup := ref.cleanupHandlers[i]
		if cleanup != nil {
			cleanup()
		}
	}
}

func (ref *ExecutionContext) WarnOn(err error) {
	if err != nil {
		log.WithError(err).Debug("error.warning")
	}
}

func (ref *ExecutionContext) FailOn(err error) {
	if err != nil {
		ref.doCleanup()

		//not using FailOn from errutil to control the flow/output better
		stackData := debug.Stack()
		log.WithError(err).WithFields(log.Fields{
			"stack": string(stackData),
		}).Error("terminating")

		if ref.Out != nil {
			ref.Out.Info("fail.on", OutVars{"version": v.Current()})
		}

		ref.exit(-1)
	}
}

func (ref *ExecutionContext) Fail(reason string) {
	ref.doCleanup()

	//not using FailOn from errutil to control the flow/output better
	stackData := debug.Stack()
	log.WithFields(log.Fields{
		"stack":  string(stackData),
		"reason": reason,
	}).Error("terminating")

	if ref.Out != nil {
		ref.Out.Info("fail.on", OutVars{"version": v.Current()})
	}

	ShowCommunityInfo(ref.Out.OutputFormat)
	ref.exit(-1)
}

func (ref *ExecutionContext) exit(exitCode int) {
	if ref.Out != nil {
		ref.Out.Info("exit", OutVars{
			"code":     exitCode,
			"version":  v.Current(),
			"location": fsutil.ExeDir(),
			"args":     os.Args})

	}

	ShowCommunityInfo(ref.Out.OutputFormat)
	os.Exit(exitCode)
}

func NewExecutionContext(
	cmdName string,
	quiet bool,
	outputFormat string,
	dataChannels ...map[string]chan interface{}) *ExecutionContext {
	var chs map[string]chan interface{}
	if len(dataChannels) > 0 {
		chs = dataChannels[0]
	}

	ref := &ExecutionContext{
		Out:  NewOutput(cmdName, quiet, outputFormat, chs),
		Args: os.Args,
	}

	return ref
}

type Output struct {
	CmdName        string
	Quiet          bool
	OutputFormat   string
	DataChannels   map[string]chan interface{}
	internalDataCh chan interface{}
}

func NewOutput(cmdName string, quiet bool, outputFormat string, channels map[string]chan interface{}) *Output {
	ref := &Output{
		CmdName:        cmdName,
		Quiet:          quiet,
		OutputFormat:   outputFormat,
		DataChannels:   channels,
		internalDataCh: make(chan interface{}),
	}

	// We want to listen to the internal channel for any data
	// And dump it onto the appropriate DataChannels
	go func() {
		for data := range ref.internalDataCh {
			log.Debugf("execontext internal data: %v\n", data)
			if data != nil {
				for _, ch := range ref.DataChannels {
					ch <- data
				}
			}
		}
	}()

	return ref
}

func NoColor() {
	color.NoColor = true
}

type OutVars map[string]interface{}

func (ref *Output) LogDump(logType string, data string, params ...OutVars) {
	if ref.Quiet {
		return
	}

	var info string
	msg := map[string]string{}
	var jsonData []byte

	msg["cmd"] = ref.CmdName
	msg["log"] = logType
	msg["data"] = data

	if len(params) > 0 {
		kvSet := params[0]
		if len(kvSet) > 0 {
			var builder strings.Builder
			for k, v := range kvSet {
				msg[k] = fmt.Sprintf("%v", v)
				builder.WriteString(kcolor(k))
				builder.WriteString("=")
				builder.WriteString(fmt.Sprintf("'%s'", vcolor("%v", v)))
				builder.WriteString(" ")
			}

			info = builder.String()
		}
	}
	switch ref.OutputFormat {
	case ofJSON:
		jsonData, _ = json.Marshal(msg)
		fmt.Println(string(jsonData))
	case ofText:
		fmt.Printf("cmd=%s log='%s' event=LOG.START %s ====================\n", ref.CmdName, logType, info)
		fmt.Println(data)
		fmt.Printf("cmd=%s log='%s' event=LOG.END %s ====================\n", ref.CmdName, logType, info)
	default:
		log.Fatalf("Unknown console output flag: %s\n. It should be either 'text' or 'json", ref.OutputFormat)
	}

}

func (ref *Output) Prompt(data string) {
	if ref.Quiet {
		return
	}

	switch ref.OutputFormat {
	case ofJSON:
		//marshal data to json
		var jsonData []byte
		if len(data) > 0 {
			msg := map[string]string{
				"cmd":    ref.CmdName,
				"prompt": data,
			}
			jsonData, _ = json.Marshal(msg)
			fmt.Println(string(jsonData))
		}
	case ofText:
		color.Set(color.FgHiRed)
		defer color.Unset()

		fmt.Printf("cmd=%s prompt='%s'\n", ref.CmdName, data)
	default:
		log.Fatalf("Unknown console output flag: %s\n. It should be either 'text' or 'json", ref.OutputFormat)
	}

}

func (ref *Output) Error(errType string, data string) {
	if ref.Quiet {
		return
	}

	switch ref.OutputFormat {
	case ofJSON:
		//marshal data to json
		var jsonData []byte
		if len(data) > 0 {
			msg := map[string]string{
				"cmd":     ref.CmdName,
				"error":   errType,
				"message": data,
			}
			jsonData, _ = json.Marshal(msg)
			fmt.Println(string(jsonData))
		}
	case ofText:
		color.Set(color.FgHiRed)
		defer color.Unset()

		fmt.Printf("cmd=%s error=%s message='%s'\n", ref.CmdName, errType, data)
	default:
		log.Fatalf("Unknown console output flag: %s\n. It should be either 'text' or 'json", ref.OutputFormat)
	}

}

func (ref *Output) Message(data string) {
	if ref.Quiet {
		return
	}

	switch ref.OutputFormat {
	case ofJSON:
		//marshal data to json
		var jsonData []byte
		if len(data) > 0 {
			msg := map[string]string{
				"cmd":     ref.CmdName,
				"message": data,
			}
			jsonData, _ = json.Marshal(msg)
			fmt.Println(string(jsonData))
		}
	case ofText:
		color.Set(color.FgHiMagenta)
		defer color.Unset()

		fmt.Printf("cmd=%s message='%s'\n", ref.CmdName, data)
	default:
		log.Fatalf("Unknown console output flag: %s\n. It should be either 'text' or 'json", ref.OutputFormat)
	}

}

func (ref *Output) Data(channelKey string, data interface{}) {
	if ch, exists := ref.DataChannels[channelKey]; exists {
		ch <- data // Send data to the corresponding channel
		fmt.Printf("Data sent to channel '%s': %v\n", channelKey, data)
	} else {
		fmt.Printf("Channel for channelKey '%s' not found\n", channelKey)
	}
}

func (ref *Output) State(state string, params ...OutVars) {
	if ref.Quiet && ref.OutputFormat != ofSubscription {
		return
	}

	var exitInfo string
	var info string
	var sep string
	msg := map[string]string{}
	var jsonData []byte
	msg["cmd"] = ref.CmdName
	msg["state"] = state

	if len(params) > 0 {
		var minCount int
		kvSet := params[0]
		if exitCode, ok := kvSet["exit.code"]; ok {
			minCount = 1
			exitInfo = fmt.Sprintf(" code=%d", exitCode)
		}

		if len(kvSet) > minCount {
			var builder strings.Builder
			sep = " "

			for k, v := range kvSet {
				if k == "exit.code" {
					continue
				}
				msg["exit.info"] = exitInfo
				msg[k] = fmt.Sprintf("%v", v)
				builder.WriteString(k)
				builder.WriteString("=")
				val := fmt.Sprintf("%v", v)
				if strings.Contains(val, " ") && !strings.HasPrefix(val, `"`) {
					val = fmt.Sprintf("\"%s\"", val)
				}

				builder.WriteString(val)
				builder.WriteString(" ")
			}

			info = builder.String()
		}
	}

	switch ref.OutputFormat {
	case ofJSON:
		jsonData, _ = json.Marshal(msg)
		fmt.Println(string(jsonData))
	case ofText:
		if state == "exited" || strings.Contains(state, "error") {
			color.Set(color.FgHiRed, color.Bold)
		} else {
			color.Set(color.FgCyan, color.Bold)
		}
		defer color.Unset()

		fmt.Printf("cmd=%s state=%s%s%s%s\n", ref.CmdName, state, exitInfo, sep, info)
	case ofSubscription:
		ref.internalDataCh <- msg // Send data to the internal channel

	default:
		log.Fatalf("Unknown console output flag: %s\n. It should be either 'text' or 'json", ref.OutputFormat)
	}
}

var (
	itcolor = color.New(color.FgMagenta, color.Bold).SprintFunc()
	kcolor  = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	vcolor  = color.New(color.FgHiBlue).SprintfFunc()
)

func (ref *Output) Info(infoType string, params ...OutVars) {
	// TODO - carry this pattern to other Output methods
	if ref.Quiet && ref.OutputFormat != ofSubscription {
		return
	}

	var data string
	var sep string
	msg := map[string]string{}
	var jsonData []byte
	msg["cmd"] = ref.CmdName
	msg["info"] = infoType

	if len(params) > 0 {
		kvSet := params[0]
		if len(kvSet) > 0 {
			var builder strings.Builder
			sep = " "

			for k, v := range kvSet {
				msg[k] = fmt.Sprintf("%v", v)
				builder.WriteString(kcolor(k))
				builder.WriteString("=")
				builder.WriteString(fmt.Sprintf("'%s'", vcolor("%v", v)))
				builder.WriteString(" ")
			}

			data = builder.String()
		}
	}

	switch ref.OutputFormat {
	case ofJSON:
		jsonData, _ = json.Marshal(msg)
		fmt.Println(string(jsonData))
	case ofText:
		fmt.Printf("cmd=%s info=%s%s%s\n", ref.CmdName, itcolor(infoType), sep, data)
	case ofSubscription:
		ref.internalDataCh <- msg // Send data to the internal channel
	default:
		log.Fatalf("Unknown console output flag: %s\n. It should be either 'text' or 'json", ref.OutputFormat)
	}

}

func ShowCommunityInfo(outputFormat string) {
	lines := []struct {
		App     string `json:"app"`
		Message string `json:"message"`
		Info    string `json:"info"`
	}{
		{
			App:     consts.AppName,
			Message: "GitHub Discussions",
			Info:    consts.CommunityDiscussions,
		},
		{
			App:     consts.AppName,
			Message: "Join the CNCF Slack channel to ask questions or to share your feedback",
			Info:    consts.CommunityCNCFSlack,
		},
		{
			App:     consts.AppName,
			Message: "Join the Discord server to ask questions or to share your feedback",
			Info:    consts.CommunityDiscord,
		},
	}

	switch outputFormat {
	case ofJSON:
		for _, v := range lines {
			jsonData, _ := json.Marshal(v)
			fmt.Println(string(jsonData))
		}
	default:
		color.Set(color.FgHiMagenta)
		defer color.Unset()

		for _, v := range lines {
			fmt.Printf("app='%s' message='%s' info='%s'\n", v.App, v.Message, v.Info)
		}
	}
}
