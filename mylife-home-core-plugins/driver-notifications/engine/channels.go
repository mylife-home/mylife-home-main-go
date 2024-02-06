package engine

type Channel interface {
	Send(title string, text string)
}

var channels = make(map[string]Channel)

func Register(key string, channel Channel) {
	channels[key] = channel
}

func Unregister(key string) {
	delete(channels, key)
}

func Send(key string, title string, text string) {
	channel := channels[key]
	if channel == nil {
		logger.Errorf("No such channel '%s'", key)
		return
	}

	channel.Send(title, text)
}
