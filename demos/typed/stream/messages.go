// The messages of the stream share the connection; each one is written as identifier followed by data.
package main

import (
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	wtypeid "github.com/malivvan/binco/wirefmt/typeid"
)

const (
	LoginDTM binco.DTM = iota + 1
	ChatDTM
	UploadDTM
)

type Login struct {
	User string
}

type Chat struct {
	From string
	Text string
	At   time.Time
}

type Upload struct {
	Name string
	Data []byte
}

type typedMessage interface {
	writeTo(w wirefmt.Writer) (n int, err error)
}

func (m Login) writeTo(w wirefmt.Writer) (n int, err error) {
	return LoginTyped.Marshal(m, w)
}

func (m Chat) writeTo(w wirefmt.Writer) (n int, err error) {
	return ChatTyped.Marshal(m, w)
}

func (m Upload) writeTo(w wirefmt.Writer) (n int, err error) {
	return UploadTyped.Marshal(m, w)
}

var (
	LoginTyped  = wtypeid.NewSerializer(LoginDTM, loginSerializer{})
	ChatTyped   = wtypeid.NewSerializer(ChatDTM, chatSerializer{})
	UploadTyped = wtypeid.NewSerializer(UploadDTM, uploadSerializer{})
)

type loginSerializer struct{}

func (loginSerializer) Marshal(v Login, w wirefmt.Writer) (n int, err error) {
	return wcomposite.String.Marshal(v.User, w)
}

func (loginSerializer) Unmarshal(r wirefmt.Reader) (v Login, n int, err error) {
	v.User, n, err = wcomposite.String.Unmarshal(r)
	return
}

func (loginSerializer) Size(v Login) (size int) {
	return wcomposite.String.Size(v.User)
}

func (loginSerializer) Skip(r wirefmt.Reader) (n int, err error) {
	return wcomposite.String.Skip(r)
}

type chatSerializer struct{}

func (chatSerializer) Marshal(v Chat, w wirefmt.Writer) (n int, err error) {
	n, err = wcomposite.String.Marshal(v.From, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wcomposite.String.Marshal(v.Text, w)
	n += n1
	if err != nil {
		return
	}
	n1, err = wscalar.TimeUnixUTC.Marshal(v.At, w)
	return n + n1, err
}

func (chatSerializer) Unmarshal(r wirefmt.Reader) (v Chat, n int, err error) {
	v.From, n, err = wcomposite.String.Unmarshal(r)
	if err != nil {
		return
	}
	var n1 int
	v.Text, n1, err = wcomposite.String.Unmarshal(r)
	n += n1
	if err != nil {
		return
	}
	v.At, n1, err = wscalar.TimeUnixUTC.Unmarshal(r)
	n += n1
	return
}

func (chatSerializer) Size(v Chat) (size int) {
	return wcomposite.String.Size(v.From) + wcomposite.String.Size(v.Text) +
		wscalar.TimeUnixUTC.Size(v.At)
}

func (chatSerializer) Skip(r wirefmt.Reader) (n int, err error) {
	n, err = wcomposite.String.Skip(r)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wcomposite.String.Skip(r)
	n += n1
	if err != nil {
		return
	}
	n1, err = wscalar.TimeUnixUTC.Skip(r)
	return n + n1, err
}

type uploadSerializer struct{}

func (uploadSerializer) Marshal(v Upload, w wirefmt.Writer) (n int, err error) {
	n, err = wcomposite.String.Marshal(v.Name, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wcomposite.ByteSlice.Marshal(v.Data, w)
	return n + n1, err
}

func (uploadSerializer) Unmarshal(r wirefmt.Reader) (v Upload, n int, err error) {
	v.Name, n, err = wcomposite.String.Unmarshal(r)
	if err != nil {
		return
	}
	var n1 int
	v.Data, n1, err = wcomposite.ByteSlice.Unmarshal(r)
	n += n1
	return
}

func (uploadSerializer) Size(v Upload) (size int) {
	return wcomposite.String.Size(v.Name) + wcomposite.ByteSlice.Size(v.Data)
}

func (uploadSerializer) Skip(r wirefmt.Reader) (n int, err error) {
	n, err = wcomposite.String.Skip(r)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wcomposite.ByteSlice.Skip(r)
	return n + n1, err
}
