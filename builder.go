package messaging

type RequestBuilder struct {
}

func (b *RequestBuilder) Build() (requester Requester, err error) {
	// TODO:
	return
}

func Builder() *RequestBuilder {
	return &RequestBuilder{}
}
