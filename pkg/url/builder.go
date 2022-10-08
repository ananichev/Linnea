package url

type URLBuilder struct {
	url string

	params map[string]string
}

func NewURLBuilder() *URLBuilder {
	return &URLBuilder{
		url:    "",
		params: make(map[string]string),
	}
}

func (u *URLBuilder) SetURL(url string) *URLBuilder {
	u.url = url

	return u
}

func (u *URLBuilder) AddParam(key, value string) *URLBuilder {
	u.params[key] = value

	return u
}

func (u *URLBuilder) Build() string {
	var final = u.url

	final += "?"

	for key, value := range u.params {
		final += key + "=" + value + "&"
	}

	return final
}