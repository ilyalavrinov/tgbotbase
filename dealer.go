package botbase

import "gopkg.in/telegram-bot-api.v4"
import "regexp"
import "log"

type ServiceMsg struct {
    stopBot bool
}

type MessageDealer interface {
    init(chan<- tgbotapi.MessageConfig, chan<- ServiceMsg)
    accept(tgbotapi.Message)
    run()
    name() string
}

type HandlerTrigger struct {
    Re *regexp.Regexp
    Cmd string
}

func (t *HandlerTrigger) canHandle(msg tgbotapi.Message) bool {
    if t.Re != nil && t.Re.MatchString(msg.Text) {
        log.Printf("Message text '%s' matched regexp '%s'", msg.Text, t.Re)
        return true
    }
    if msg.IsCommand() && t.Cmd == msg.Command() {
        log.Printf("Message text '%s' matched command '%s'", msg.Text, t.Cmd)
        return true
    }
    log.Printf("Message text '%s' doesn't match either command '%s' or regexp '%s'", msg.Text, t.Cmd, t.Re)
    return false
}

type IncomingMessageHandler interface {
    Init(chan<- tgbotapi.MessageConfig, chan<- ServiceMsg) HandlerTrigger
    HandleOne(tgbotapi.Message)
    Name() string
}

type IncomingMessageDealer struct {
    handler IncomingMessageHandler
    trigger HandlerTrigger
    inMsgCh chan tgbotapi.Message
}

func NewIncomingMessageDealer(h IncomingMessageHandler) *IncomingMessageDealer {
    d := &IncomingMessageDealer{handler: h}
    return d
}

func (d *IncomingMessageDealer) init(outMsgCh chan<- tgbotapi.MessageConfig, srvCh chan<- ServiceMsg) {
    d.trigger = d.handler.Init(outMsgCh, srvCh)
    d.inMsgCh = make(chan tgbotapi.Message, 0)
}

func (d *IncomingMessageDealer) accept(msg tgbotapi.Message) {
    if d.trigger.canHandle(msg) {
        d.inMsgCh<- msg
    }
}

func (d *IncomingMessageDealer) run() {
    go func() {
        for msg := range d.inMsgCh {
            d.handler.HandleOne(msg)
        }
    }()
}

func (d *IncomingMessageDealer) name() string {
    return d.handler.Name()
}


type BaseHandler struct {
    OutMsgCh chan<- tgbotapi.MessageConfig
    SrvCh chan<- ServiceMsg
}


type BackgroundMessageDealer struct {
    MessageDealer
    BaseHandler
}

func (d *BackgroundMessageDealer) init(outMsgCh chan<- tgbotapi.MessageConfig, srvCh chan<- ServiceMsg) {
    d.OutMsgCh = outMsgCh
    d.SrvCh = srvCh
}

func (d *BackgroundMessageDealer) accept(tgbotapi.Message) {
    // doing nothing
}

// MessageDealer::run to be overwritten by concrete implementations