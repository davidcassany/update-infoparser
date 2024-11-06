package types

import (
	"encoding/xml"
	"net/url"
	"strconv"
	"time"
)

type UpdateInfo struct {
	XMLName xml.Name `xml:"updates"`
	Updates []Update `xml:"update"`
}

type Update struct {
	XMLName     xml.Name    `xml:"update"`
	Type        string      `xml:"type,attr"`
	Status      string      `xml:"status,attr"`
	ID          string      `xml:"id"`
	Title       string      `xml:"title"`
	Severity    string      `xml:"severity"`
	Release     string      `xml:"release"`
	Issued      Issued      `xml:"issued"`
	References  []Reference `xml:"references>reference"`
	Description string      `xml:"description"`
	Packages    []Package   `xml:"pkglist>collection>package"`
}

type date time.Time

type Issued struct {
	XMLName xml.Name `xml:"issued"`
	Date    *date    `xml:"date,attr"`
}

func (d date) String() string {
	return time.Time(d).String()
}

func (d *date) UnmarshalXMLAttr(attr xml.Attr) error {
	ts, err := strconv.ParseInt(attr.Value, 10, 64)
	if err != nil {
		return err
	}
	t := time.Unix(ts, 0)
	*d = date(t)

	return nil
}

type href url.URL

type Reference struct {
	XMLName xml.Name `xml:"reference"`
	URL     href     `xml:"href,attr"`
	ID      string   `xml:"id,attr"`
	Title   string   `xml:"title,attr"`
	Type    string   `xml:"type,attr"`
}

func (h href) String() string {
	u := url.URL(h)
	return u.String()
}

func (h *href) UnmarshalXMLAttr(attr xml.Attr) error {
	u, err := url.Parse(attr.Value)
	if err != nil {
		return err
	}
	*h = href(*u)
	return nil
}

type Package struct {
	XMLName  xml.Name `xml:"package"`
	Name     string   `xml:"name,attr"`
	Version  string   `xml:"version,attr"`
	Release  string   `xml:"release,attr"`
	Arch     string   `xml:"arch,attr"`
	Filename string   `xml:"filename"`
}
