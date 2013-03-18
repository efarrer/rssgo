// Copyright 2012 Evan Farrer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/* Package rssgo provides a basic interface for processing RSS version 2.0 feeds
   as defined by http://cyber.law.harvard.edu/rss/rss.html
*/
package rssgo

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var allowableLanguageMap map[string]bool
var allowableCloudProtocolMap map[string]bool
var allowableSkipDays map[string]bool

// The version of RSS that rssgo supports. Suitable for use as the
// rss>channel>version value
const Version = "2.0"

// The URL of the RSS 2.0 spec. Suitable for use as the Rss.Docs value
const DocsURL = "http://blogs.law.harvard.edu/tech/rss"

// The default image width value
const DefaultWidth = 88

// The default image height value
const DefaultHeight = 31

type Rss struct {
	XMLName string `xml:"rss"`

	// Required. Value should be rssgo.Version.
	Version string `xml:"version,attr"`

	// Required. The title of your channel.
	Title string `xml:"channel>title"`

	// Required. The URL of your website.
	Link string `xml:"channel>link"`

	// Required. The description of the channel.
	Description string `xml:"channel>description"`

	// Optional. If present allowable values are found at
	// http://cyber.law.harvard.edu/rss/languages.html
	Language string `xml:"channel>language,omitempty"`

	// Optional. Copyright notice for channel content
	Copyright string `xml:"channel>copyright,omitempty"`

	// Optional. Email address for the managing editor
	ManagingEditor string `xml:"channel>managingEditor,omitempty"`

	// Optional. Email address for the channel's web master
	WebMaster string `xml:"channel>webMaster,omitempty"`

	// Optional. Publication date of the channel. See rssgo.ComposeRssDate and
	// rssgo.ParseRssDate
	PubDate string `xml:"channel>pubDate,omitempty"`

	// Optional. Date of last change to the channel content. See
	// rssgo.ComposeRssDate and rssgo.ParseRssDate
	LastBuildDate string `xml:"channel>lastBuildDate,omitempty"`

	// Optional. The hierarchical categorizations.
	Categories []Category `xml:"channel>category"`

	// Optional. The program used to generate the RSS
	Generator string `xml:"channel>generator,omitempty"`

	// Optional. The URL for the document describing the RSS format. Should be
	// the DocsURL constant
	Docs string `xml:"channel>docs,omitempty"`

	// Optional. A web service that supports the rssCloud interface
	Cloud *Cloud `xml:"channel>cloud"`

	// Optional. The number of minutes the channel can be cached
	Ttl int `xml:"channel>ttl,omitempty"`

	// Optional. An image that represents the channel.
	Image *Image `xml:"channel>image"`

	// Optional. The PICS rating for this channel. See http://www.w3.org/PICS/
	Rating string `xml:"channel>rating,omitempty"`

	// Optional. The channel's text input box
	TextInput *TextInput `xml:"channel>textInput"`

	// Optional. The hours when aggregators may not read the channel
	SkipHours *Hours `xml:"channel>skipHours,omitempty"`

	// Optional. The days when aggregators may not read the channel
	SkipDays *Days `xml:"channel>skipDays,omitempty"`

	// Optional. The RSS feed's items
	Items []Item `xml:"channel>item"`
}

// A RSS feeds item
type Item struct {
	// Either the title or the description are required. The title of the item.
	Title string `xml:"title,omitempty"`

	// Optional. The URL of the item
	Link string `xml:"link,omitempty"`

	// Either the title or the description are required. The item description.
	Description string `xml:"description,omitempty"`

	// Optional. The authors email address
	Author string `xml:"author,omitempty"`

	// Optional. The items hierarchical categorizations.
	Categories []Category `xml:"category"`

	// Optional. The URL for the page containing the items comments.
	Comments string `xml:"comments,omitempty"`

	// Optional. A media object attached to the item.
	Enclosure *Enclosure `xml:"enclosure"`

	// Optional. A unique identifier for the item
	Guid *Guid `xml:"guid"`

	// Optional. Publication date of the item. See rssgo.ComposeRssDate and
	// rssgo.ParseRssDate
	PubDate string `xml:"pubDate,omitempty"`

	// Optional. The RSS channel the item came from.
	Source *Source `xml:"source"`
}

// The RSS channel the item came from.
type Source struct {
	// Required. The title of the channel where the item came from.
	Source string `xml:",chardata"`

	// Required. The URL of the channel where the item came from.
	Url string `xml:"url,attr"`
}

// A unique identifier for the item
type Guid struct {

	// Required. The items GUID
	Guid string `xml:",chardata"`

	// Optional. If set to true the Guid must be a URL
	IsPermaLink bool `xml:"isPermaLink,attr,omitempty"`
}

// A media object for an item
type Enclosure struct {
	// Required. The enclosures URL.
	Url string `xml:"url,attr"`

	// Required. The enclosures size.
	Length int64 `xml:"length,attr,omitempty"`

	// Required. The enclosures MIME type.
	Type string `xml:"type,attr"`
}

// A day when an aggregator may not read the channel
type Days struct {
	// Required. The day
	Days []string `xml:"day"`
}

// An hour when an aggregator may not read the channel
type Hours struct {
	// Required. The hour
	Hours []int `xml:"hour"`
}

// The Rsa channel's text input box
type TextInput struct {
	// Required. The text input's Submit button text
	Title string `xml:"title,omitempty"`

	// Required. The text input's desxcription.
	Description string `xml:"description,omitempty"`

	// Required. The text input objects name
	Name string `xml:"name,omitempty"`

	// Required. The URL of the CGI script that processes the text input request.
	Link string `xml:"link,omitempty"`
}

// An RSS channel's image
type Image struct {
	// Required. The URL to the GIF, JPEG, or PNG image
	Url string `xml:"url"`

	// Required. The image title (should probably match the channels title)
	Title string `xml:"title"`

	// Required. The image link (should probably match the channels link)
	Link string `xml:"link"`

	// Optional. The image width. 
	// Note: If the element is missing from the XML this field will have a value
	// of 0. The field value should be treated as having a value of DefaultWidth
	Width int `xml:"width,omitempty"`

	// Optional. The image height. 
	// Note: If the element is missing from the XML this field will have a value
	// of 0. The field value should be treated as having a value of DefaultHeight
	Height int `xml:"height,omitempty"`
}

// The rssCloud interface parameters
type Cloud struct {
	// Required. The rssCloud domain
	Domain string `xml:"domain,attr"`

	// Required. The rssCloud port 0-65535
	Port int `xml:"port,attr,omitempty"`

	// Required. The rssCloud path
	Path string `xml:"path,attr"`

	// Required. The name of the rssCloud register procedure
	RegisterProcedure string `xml:"registerProcedure,attr"`

	// Required. The protocol. Must be xml-rpc, soap, or http-post
	Protocol string `xml:"protocol,attr"`
}

// A hierarchical categorization type
type Category struct {
	// Required. A hierarchical categorizations
	Category string `xml:",chardata"`

	// Optional. The domain URL
	Domain string `xml:"domain,attr,omitempty"`
}

// Verifies that the contents of the Rss object will conform to the RSS 2.0
// spec.
func Verify(r *Rss) error {

	if r.Version != Version {
		return errors.New(fmt.Sprintf("Bad version. Expecting %v", Version))
	}

	if r.Title == "" {
		return errors.New("Empty title. The title must be set")
	}

	_, err := url.Parse(r.Link)
	if err != nil {
		return errors.New(fmt.Sprintf("Bad channel link. Expecting a valid URL (%v)", err))
	}

	if r.Description == "" {
		return errors.New("Empty description. The description must be set")
	}

	if r.Language != "" && !allowableLanguageMap[r.Language] {
		return errors.New(`Invalid language. Allowable language values are found 
at http://cyber.law.harvard.edu/rss/languages.html`)
	}

	// Verify the validity of field dates
	verifyDateFields := func(field string) error {
		if field != "" {
			_, err := ParseRssDate(field)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if err := verifyDateFields(r.PubDate); err != nil {
		return errors.New(fmt.Sprintf("Unable to parse the RSS PubDate (%v)", err))
	}

	if err := verifyDateFields(r.LastBuildDate); err != nil {
		return errors.New(fmt.Sprintf("Unable to parse the RSS LastBuildDate (%v)", err))
	}

	for i := 0; i != len(r.Categories); i++ {
		if r.Categories[i].Category == "" {
			return errors.New("Category should not be empty.")
		}
	}

	if r.Docs != "" && r.Docs != DocsURL {
		return errors.New(fmt.Sprintf("Docs should be empty or %v", DocsURL))
	}

	if r.Cloud != nil {
		if r.Cloud.Domain == "" {
			return errors.New("Cloud domain must not be empty")
		}
		if err != nil || r.Cloud.Port < 1 || r.Cloud.Port > 65535 {
			return errors.New("Cloud port must be from 1 to 65535.")
		}
		if r.Cloud.Path == "" || r.Cloud.Path[0] != '/' {
			return errors.New("Invalid cloud path.")
		}
		if r.Cloud.RegisterProcedure == "" {
			return errors.New("Invalid cloud register procedure.")
		}
		if !allowableCloudProtocolMap[r.Cloud.Protocol] {
			return errors.New("Invalid cloud protocol. It must be xml-rpc, soap, or http-post")
		}
	}

	if r.Ttl < 0 {
		return errors.New("Ttl field must be a positive integer.")
	}

	if r.Image != nil {
		_, err := url.Parse(r.Image.Url)
		if err != nil {
			return errors.New(fmt.Sprintf("Bad image url. Expecting a valid URL (%v)", err))
		}

		if r.Image.Title == "" {
			return errors.New("Empty image title. The image title must be set")
		}

		_, err = url.Parse(r.Image.Link)
		if err != nil {
			return errors.New(fmt.Sprintf("Bad image link. Expecting a valid URL (%v)", err))
		}

		if r.Image.Width < 0 || r.Image.Width > 144 {
			return errors.New("Image width must be from 1 to 144.")
		}

		if r.Image.Height < 0 || r.Image.Height > 400 {
			return errors.New("Image heigth must be from 1 to 400.")
		}
	}

	if r.TextInput != nil {
		if r.TextInput.Title == "" {
			return errors.New("Text input's title must be set.")
		}

		if r.TextInput.Description == "" {
			return errors.New("Text input's description must be set.")
		}

		if r.TextInput.Name == "" {
			return errors.New("Text input's name must be set.")
		}

		_, err := url.Parse(r.TextInput.Link)
		if err != nil {
			return errors.New(fmt.Sprintf("Bad text input's link. Expecting a valid URL (%v)", err))
		}
	}

	if r.SkipHours != nil {
		for h := 0; h != len(r.SkipHours.Hours); h++ {
			hour := r.SkipHours.Hours[h]
			if err != nil || hour < 0 || hour > 23 {
				return errors.New("The skipHour's hour must be from 0 to 23")
			}
		}
	}

	if r.SkipDays != nil {
		for d := 0; d != len(r.SkipDays.Days); d++ {
			if !allowableSkipDays[r.SkipDays.Days[d]] {
				return errors.New("Invalid skip day. Allowable skip days can be found at http://cyber.law.harvard.edu/rss/skipHoursDays.html#skiphours")
			}
		}
	}

	for i := 0; i != len(r.Items); i++ {
		if r.Items[i].Title == "" {
			if r.Items[i].Description == "" {
				return errors.New("The item title or description must be set.")
			}
		}

		if r.Items[i].Link != "" {
			_, err := url.Parse(r.Items[i].Link)
			if err != nil {
				return errors.New(fmt.Sprintf("Bad item link. Expecting a valid URL (%v)", err))
			}
		}

		if r.Items[i].Comments != "" {
			_, err := url.Parse(r.Items[i].Comments)
			if err != nil {
				return errors.New(fmt.Sprintf("Bad item comments. Expecting a valid URL (%v)", err))
			}
		}

		if r.Items[i].Enclosure != nil {
			_, err := url.Parse(r.Items[i].Enclosure.Url)
			if err != nil {
				return errors.New(fmt.Sprintf("Bad item enclosure url. Expecting a valid URL (%v)", err))
			}

			if r.Items[i].Enclosure.Length <= 0 {
				return errors.New("The item enclosure length should not be greater than zero.")
			}

			if r.Items[i].Enclosure.Type == "" {
				return errors.New("The item enclosure type must be set.")
			}
		}

		if r.Items[i].Guid != nil {
			if r.Items[i].Guid.IsPermaLink {
				_, err := url.Parse(r.Items[i].Guid.Guid)
				if err != nil {
					return errors.New(fmt.Sprintf("Bad item guid body. Expecting a valid URL (%v)", err))
				}
			}
		}

		if err := verifyDateFields(r.Items[i].PubDate); err != nil {
			return errors.New(fmt.Sprintf("Unable to parse the item PubDate (%v)", err))
		}

		if r.Items[i].Source != nil {
			if r.Items[i].Source.Source == "" {
				return errors.New("The item source must be set.")
			}

			_, err := url.Parse(r.Items[i].Source.Url)
			if err != nil {
				return errors.New(fmt.Sprintf("Bad item source url. Expecting a valid URL (%v)", err))
			}
		}
	}

	return nil
}

const dayPrefix = "Mon, "
const dayMonth = "02 Jan "
const fourYear = "2006 "
const twoYear = "06 "
const includeSeconds = "15:04:05 "
const excludeSeconds = "15:04 "
const localDifferential = "-0700"
const zone = "MST"
const rfc822WithFourCharacterYear = "02 Jan 2006 15:04 MST"

/*
 Parses a date/time string that matches the RSS 2.0 format (RFC822 with 2 or 4
 character year) into a time.Time type.
*/
func ParseRssDate(date string) (time.Time, error) {

	format := ""

	// See if it has a leading day
	if strings.Contains(date, ",") {
		format += dayPrefix
	}

	format += dayMonth

	// See if it has a 4 character year
	if matched, _ := regexp.Match(".*[A-Z][a-z]{2} [0-9]{4}", []byte(date)); matched {
		format += fourYear
	} else {
		format += twoYear
	}

	if 2 == strings.Count(date, ":") {
		format += includeSeconds
	} else {
		format += excludeSeconds
	}

	if strings.Contains(date, "+") || strings.Contains(date, "-") {
		format += localDifferential
	} else {
		format += zone
	}

	return time.Parse(format, date)
}

/*
 Compose a date/time string that matches the RSS 2.0 format (RFC822 with 4
 character year) into a time.Time type.
*/
func ComposeRssDate(date time.Time) string {
	return date.Format(rfc822WithFourCharacterYear)
}

func init() {
	allowableCloudProtocolMap = map[string]bool{
		"xml-rpc":   true,
		"soap":      true,
		"http-post": true}
	allowableSkipDays = map[string]bool{
		"Monday":    true,
		"Tuesday":   true,
		"Wednesday": true,
		"Thursday":  true,
		"Friday":    true,
		"Saturday":  true,
		"Sunday":    true,
	}
	allowableLanguageMap = map[string]bool{
		"af":    true,
		"sq":    true,
		"eu":    true,
		"be":    true,
		"bg":    true,
		"ca":    true,
		"zh-cn": true,
		"zh-tw": true,
		"hr":    true,
		"cs":    true,
		"da":    true,
		"nl":    true,
		"nl-be": true,
		"nl-nl": true,
		"en":    true,
		"en-au": true,
		"en-bz": true,
		"en-ca": true,
		"en-ie": true,
		"en-jm": true,
		"en-nz": true,
		"en-ph": true,
		"en-za": true,
		"en-tt": true,
		"en-gb": true,
		"en-us": true,
		"en-zw": true,
		"et":    true,
		"fo":    true,
		"fi":    true,
		"fr":    true,
		"fr-be": true,
		"fr-ca": true,
		"fr-fr": true,
		"fr-lu": true,
		"fr-mc": true,
		"fr-ch": true,
		"gl":    true,
		"gd":    true,
		"de":    true,
		"de-at": true,
		"de-de": true,
		"de-li": true,
		"de-lu": true,
		"de-ch": true,
		"el":    true,
		"haw":   true,
		"hu":    true,
		"is":    true,
		"in":    true,
		"ga":    true,
		"it":    true,
		"it-it": true,
		"it-ch": true,
		"ja":    true,
		"ko":    true,
		"mk":    true,
		"no":    true,
		"pl":    true,
		"pt":    true,
		"pt-br": true,
		"pt-pt": true,
		"ro":    true,
		"ro-mo": true,
		"ro-ro": true,
		"ru":    true,
		"ru-mo": true,
		"ru-ru": true,
		"sr":    true,
		"sk":    true,
		"sl":    true,
		"es":    true,
		"es-ar": true,
		"es-bo": true,
		"es-cl": true,
		"es-co": true,
		"es-cr": true,
		"es-do": true,
		"es-ec": true,
		"es-sv": true,
		"es-gt": true,
		"es-hn": true,
		"es-mx": true,
		"es-ni": true,
		"es-pa": true,
		"es-py": true,
		"es-pe": true,
		"es-pr": true,
		"es-es": true,
		"es-uy": true,
		"es-ve": true,
		"sv":    true,
		"sv-fi": true,
		"sv-se": true,
		"tr":    true,
		"uk":    true}
}
