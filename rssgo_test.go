// Copyright 2012 Evan Farrer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rssgo

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestComposeRssDate(t *testing.T) {

	then := time.Date(1974, time.July, 23, 9, 10, 11, 12, time.UTC)
	expected := "23 Jul 1974 09:10 UTC"
	actual := ComposeRssDate(then)
	if expected != actual {
		t.Fatalf("ComposeRssDate returned incorrect date/time string expected: %v got: %v\n",
			expected, actual)
	}
}

func TestParseRssDate(t *testing.T) {

	testString := func(str string, expected time.Time) {
		actual, err := ParseRssDate(str)
		if err != nil {
			t.Fatalf("Unexpected error (%v) when parsing %v\n", err, str)
		}
		if actual != expected {
			t.Fatalf("Unexpected time.Time when parsing %v. Expected: %v got: %v\n",
				str, expected, actual)
		}
	}

	expected := time.Date(1974, time.July, 23, 9, 10, 0, 0, time.UTC)
	testString("23 Jul 74 09:10 UTC", expected)
	testString("23 Jul 1974 09:10 UTC", expected)
	testString("Wed, 23 Jul 74 09:10 UTC", expected)
	testString("Wed, 23 Jul 1974 09:10 UTC", expected)

	expected = time.Date(1974, time.July, 23, 9, 10, 30, 0, time.UTC)
	testString("Wed, 23 Jul 1974 09:10:30 UTC", expected)

	str := "Wed, 23 Jul 1974 09:10:30 +0700"
	actual, err := ParseRssDate(str)
	if err != nil {
		t.Fatalf("Unexpected error (%v) when parsing %v\n", err, str)
	}
	// For numeric location just copy it
	expected = time.Date(1974, time.July, 23, 9, 10, 30, 0, actual.Location())
	if actual != expected {
		t.Fatalf("Unexpected time.Time when parsing %v. Expected: %#v got: %#v\n",
			str, expected, actual)
	}
}

func TestVerify(t *testing.T) {

	// Function for creating a valid Rss struct
	createValidRss := func() *Rss {
		return &Rss{Version: Version,
			Title:       "title",
			Link:        "http://github.com/efarrer/rssgo/",
			Description: "A podcast"}
	}

	verifyShouldFail := func(r *Rss, reason string) {
		if err := Verify(r); err == nil {
			t.Fatalf("Verify should fail: %v\n", reason)
		}
	}
	verifyShouldPass := func(r *Rss, testing string) {
		if err := Verify(r); err != nil {
			t.Fatalf("Verify should pass: %v %v\n", testing, err)
		}
	}

	// Basic should pass
	rss := createValidRss()
	verifyShouldPass(rss, "All good")

	// Version
	rss = createValidRss()
	rss.Version = ""
	verifyShouldFail(rss, "missing version")

	rss = createValidRss()
	rss.Version = "0.0"
	verifyShouldFail(rss, "incorrect version")

	rss = createValidRss()
	rss.Version = Version
	verifyShouldPass(rss, "Correct version")

	// Title
	rss = createValidRss()
	rss.Title = ""
	verifyShouldFail(rss, "missing title")

	rss = createValidRss()
	rss.Title = "title"
	verifyShouldPass(rss, "Correct title")

	// Link
	rss = createValidRss()
	rss.Link = ""
	verifyShouldFail(rss, "missing link")

	rss = createValidRss()
	rss.Link = "##This is a bad link!"
	verifyShouldFail(rss, "malformed link")

	rss = createValidRss()
	rss.Link = "http://foo.com"
	verifyShouldPass(rss, "Correct link")

	// Description
	rss = createValidRss()
	rss.Description = ""
	verifyShouldFail(rss, "missing description")

	rss = createValidRss()
	rss.Description = "Any description will work"
	verifyShouldPass(rss, "Correct description")

	// Language
	rss = createValidRss()
	rss.Language = ""
	verifyShouldPass(rss, "Language can be empty")

	rss = createValidRss()
	rss.Language = "en-us"
	verifyShouldPass(rss, "Correct language")

	rss = createValidRss()
	rss.Language = "pig-latin"
	verifyShouldFail(rss, "Malformed language")

	// PubDate
	rss = createValidRss()
	rss.PubDate = ""
	verifyShouldPass(rss, "PubDate can be empty")

	rss = createValidRss()
	rss.PubDate = ComposeRssDate(time.Now())
	verifyShouldPass(rss, "Valid PubDate")

	rss = createValidRss()
	rss.PubDate = "Some time tomorrow"
	verifyShouldFail(rss, "Invalid PubDate")

	// LastBuildDate
	rss = createValidRss()
	rss.LastBuildDate = ""
	verifyShouldPass(rss, "LastBuildDate can be empty")

	rss = createValidRss()
	rss.LastBuildDate = ComposeRssDate(time.Now())
	verifyShouldPass(rss, "Valid LastBuildDate")

	rss = createValidRss()
	rss.LastBuildDate = "Some time tomorrow"
	verifyShouldFail(rss, "Invalid LastBuildDate")

	// Categories
	rss = createValidRss()
	rss.Categories = []Category{}
	verifyShouldPass(rss, "Can have empty categories")

	rss = createValidRss()
	rss.Categories = []Category{{"foo", "http://www.example.com"}}
	verifyShouldPass(rss, "Can have more than 0 entries")

	rss = createValidRss()
	rss.Categories = []Category{{"", "http://www.example.com"}}
	verifyShouldFail(rss, "Category should not be empty")

	rss = createValidRss()
	rss.Categories = []Category{{"foo", "#http://www.example.com"}}
	verifyShouldPass(rss, "Domain can be a URL or a free form string")

	// Docs
	rss = createValidRss()
	rss.Docs = ""
	verifyShouldPass(rss, "Can have empty docs")

	rss = createValidRss()
	rss.Docs = DocsURL
	verifyShouldPass(rss, "Can be the DocsURL constant")

	rss = createValidRss()
	rss.Docs = "stuff"
	verifyShouldFail(rss, "Docs should be empty or DocsURL")

	// Cloud
	rss = createValidRss()
	rss.Cloud = nil
	verifyShouldPass(rss, "Can have empty cloud")

	createValidCloud := func() *Cloud {
		return &Cloud{"example.com", 80, "/path", "foo.bar", "xml-rpc"}
	}
	rss = createValidRss()
	rss.Cloud = createValidCloud()
	verifyShouldPass(rss, "Valid cloud")

	rss = createValidRss()
	rss.Cloud = createValidCloud()
	rss.Cloud.Domain = ""
	verifyShouldFail(rss, "Domain must not be empty")

	rss = createValidRss()
	rss.Cloud = createValidCloud()
	rss.Cloud.Port = 0
	verifyShouldFail(rss, "Port must be 1-65535")

	rss = createValidRss()
	rss.Cloud = createValidCloud()
	rss.Cloud.Port = 65536
	verifyShouldFail(rss, "Port must be 1-65535")

	rss = createValidRss()
	rss.Cloud = createValidCloud()
	rss.Cloud.Path = ""
	verifyShouldFail(rss, "Path must not be empty")

	rss = createValidRss()
	rss.Cloud = createValidCloud()
	rss.Cloud.Path = "foo"
	verifyShouldFail(rss, "Path must start with '/'")

	rss = createValidRss()
	rss.Cloud = createValidCloud()
	rss.Cloud.RegisterProcedure = ""
	verifyShouldFail(rss, "RegisterProcedure must not be empty")

	rss = createValidRss()
	rss.Cloud = createValidCloud()
	rss.Cloud.Protocol = ""
	verifyShouldFail(rss, "Protocol must be xml-rpc, soap or http-post")

	// Ttl
	rss = createValidRss()
	verifyShouldPass(rss, "Can have empty ttl")

	rss = createValidRss()
	rss.Ttl = 60
	verifyShouldPass(rss, "Can have number")

	rss = createValidRss()
	rss.Ttl = -1
	verifyShouldFail(rss, "Must be a positive number")

	// Image
	rss = createValidRss()
	rss.Image = nil
	verifyShouldPass(rss, "Can have empty image")

	createValidImage := func() *Image {
		return &Image{Url: "http://www.image.com/image.png", Title: "title",
			Link: "http://link.com"}
	}
	rss = createValidRss()
	rss.Image = createValidImage()
	verifyShouldPass(rss, "Can have basic image")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Url = ""
	verifyShouldFail(rss, "Image url must be set")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Url = "#httpsdf;as/"
	verifyShouldFail(rss, "Image url must be valid")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Title = ""
	verifyShouldFail(rss, "Image title must be set")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Link = ""
	verifyShouldFail(rss, "Image link must be set")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Link = "#hsdaf asdfa/sf"
	verifyShouldFail(rss, "Image link must be valid")

	rss = createValidRss()
	rss.Image = createValidImage()
	verifyShouldPass(rss, "Image width can be empty")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Width = 10
	verifyShouldPass(rss, "Image width can be numeric")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Width = 145
	verifyShouldFail(rss, "Image width should be < 145")

	rss = createValidRss()
	rss.Image = createValidImage()
	verifyShouldPass(rss, "Image height can be empty")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Height = 30
	verifyShouldPass(rss, "Image height can be set")

	rss = createValidRss()
	rss.Image = createValidImage()
	rss.Image.Height = 401
	verifyShouldFail(rss, "Image height should be < 401")

	// Rating
	rss = createValidRss()
	verifyShouldPass(rss, "Rating can be empty")

	rss = createValidRss()
	rss.Rating = "some rating"
	verifyShouldPass(rss, "Rating can be set")

	// TextInput
	rss = createValidRss()
	rss.TextInput = nil
	verifyShouldPass(rss, "Text input can be nil")

	createTextInput := func() *TextInput {
		return &TextInput{"title", "the text input", "textinput",
			"http://www.foo.com"}
	}
	rss = createValidRss()
	rss.TextInput = createTextInput()
	verifyShouldPass(rss, "Test valid text input")

	rss = createValidRss()
	rss.TextInput = createTextInput()
	rss.TextInput.Title = ""
	verifyShouldFail(rss, "Text input title is required")

	rss = createValidRss()
	rss.TextInput = createTextInput()
	rss.TextInput.Description = ""
	verifyShouldFail(rss, "Text input description is required")

	rss = createValidRss()
	rss.TextInput = createTextInput()
	rss.TextInput.Name = ""
	verifyShouldFail(rss, "Text input name is required")

	rss = createValidRss()
	rss.TextInput = createTextInput()
	rss.TextInput.Link = ""
	verifyShouldFail(rss, "Text input link is required")

	rss = createValidRss()
	rss.TextInput = createTextInput()
	rss.TextInput.Link = "#http:/sdfa as dfasl.com"
	verifyShouldFail(rss, "Text input link must be an URL")

	// SkipHours
	rss = createValidRss()
	rss.SkipHours = &Hours{[]int{}}
	verifyShouldPass(rss, "Skip hours can be empty")

	createValidSkipHours := func() *Hours {
		return &Hours{[]int{0, 1}}
	}

	rss = createValidRss()
	rss.SkipHours = createValidSkipHours()
	verifyShouldPass(rss, "Skip hours can be set")

	rss = createValidRss()
	rss.SkipHours = createValidSkipHours()
	rss.SkipHours.Hours = []int{-1}
	verifyShouldFail(rss, "An hour must be > 0")

	rss = createValidRss()
	rss.SkipHours = createValidSkipHours()
	rss.SkipHours.Hours = []int{24}
	verifyShouldFail(rss, "An hour must be < 24")

	// SkipDays
	rss = createValidRss()
	rss.SkipDays = &Days{}
	verifyShouldPass(rss, "Skip days can be empty")

	createValidSkipDays := func() *Days {
		return &Days{[]string{"Monday", "Tuesday"}}
	}

	rss = createValidRss()
	rss.SkipDays = createValidSkipDays()
	verifyShouldPass(rss, "Skip days can be set")

	rss = createValidRss()
	rss.SkipDays = createValidSkipDays()
	rss.SkipDays.Days[0] = ""
	verifyShouldFail(rss, "An skip day must be set")

	rss = createValidRss()
	rss.SkipDays = createValidSkipDays()
	rss.SkipDays.Days[0] = "somday"
	verifyShouldFail(rss, "An skip day must be valid")

	// Items
	rss = createValidRss()
	rss.Items = []Item{}
	verifyShouldPass(rss, "Items days can be empty")

	createValidItems := func() []Item {
		return []Item{{"title", "http://link.com", "the item", "author@authors.com",
			[]Category{{"categories", ""}}, "http://comments.com", nil, nil,
			"23 Jul 74 09:10 UTC", nil},
			{"title2", "http://link2.com", "the 2 item", "author2@authors.com",
				[]Category{}, "http://comments2.com", nil, nil, "23 Jul 74 08:10 UTC",
				nil}}
	}

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Title = ""
	verifyShouldPass(rss, "Title or description must be set")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Description = ""
	verifyShouldPass(rss, "Title or description must be set")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Title = ""
	rss.Items[0].Description = ""
	verifyShouldFail(rss, "Title or description must be set")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Link = ""
	verifyShouldPass(rss, "Link can be empty")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Link = "#http masdf"
	verifyShouldFail(rss, "Link must be a URL")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Author = ""
	verifyShouldPass(rss, "Author can be empty")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Comments = ""
	verifyShouldPass(rss, "Comments can be empty")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Comments = "#hsdf sdf"
	verifyShouldFail(rss, "Comments must be a URL")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Enclosure = nil
	verifyShouldPass(rss, "Enclosure can be empty")

	createValidEnclosure := func() *Enclosure {
		return &Enclosure{"http://enclosure/music.mp3", 10000, "audio/mpeg"}
	}
	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Enclosure = createValidEnclosure()
	verifyShouldPass(rss, "A valid enclosure")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Enclosure = createValidEnclosure()
	rss.Items[0].Enclosure.Url = ""
	verifyShouldFail(rss, "The enclosure URL must be set")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Enclosure = createValidEnclosure()
	rss.Items[0].Enclosure.Url = "#hsdf  sdf"
	verifyShouldFail(rss, "The enclosure URL must be a valid URL")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Enclosure = createValidEnclosure()
	rss.Items[0].Enclosure.Length = 0
	verifyShouldFail(rss, "The enclosure length must non zero")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Enclosure = createValidEnclosure()
	rss.Items[0].Enclosure.Length = -1
	verifyShouldFail(rss, "The enclosure length must be > 0")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Enclosure = createValidEnclosure()
	rss.Items[0].Enclosure.Type = ""
	verifyShouldFail(rss, "The enclosure type must be set")

	// Guid
	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Guid = nil
	verifyShouldPass(rss, "The item guid can be empty")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Guid = &Guid{Guid: "guid"}
	verifyShouldPass(rss, "The item guid can be set, IsPermaLink can be empty")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Guid = &Guid{Guid: ""}
	verifyShouldPass(rss, "The item guid can't be empty")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Guid = &Guid{"#guid", true}
	verifyShouldFail(rss, "If IsPermaLink is true guid must be URL")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Guid = &Guid{"http://guid.com", true}
	verifyShouldPass(rss, "If IsPermaLink is true guid must be URL")

	// Item.PubDate
	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].PubDate = ""
	verifyShouldPass(rss, "PubDate can be empty")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].PubDate = ComposeRssDate(time.Now())
	verifyShouldPass(rss, "Valid PubDate")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].PubDate = "Some time tomorrow"
	verifyShouldFail(rss, "Invalid PubDate")

	// Source
	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Source = nil
	verifyShouldPass(rss, "Source can be empty")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Source = &Source{"title", "http://source.com"}
	verifyShouldPass(rss, "Valid source")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Source = &Source{"", "http://source.com"}
	verifyShouldFail(rss, "Source must have a body")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Source = &Source{"title", ""}
	verifyShouldFail(rss, "Source must be set")

	rss = createValidRss()
	rss.Items = createValidItems()
	rss.Items[0].Source = &Source{"title", "# sadsadf asf"}
	verifyShouldFail(rss, "Source must be a valid URL")
}

func TestSerialize(t *testing.T) {

	// Can serialize minimum
	rss := &Rss{Version: Version,
		Title:       "title",
		Link:        "http://github.com/efarrer/rssgo/",
		Description: "A podcast"}
	err := Verify(rss)
	if err != nil {
		t.Fatalf("Unable to verify minimum %v\n", err)
	}
	_, err = xml.MarshalIndent(rss, "", "    ")
	if err != nil {
		t.Fatalf("Unable to marshal minimum %v\n", err)
	}

	// Can serialize full
	rss = &Rss{Version: Version,
		Title:          "Title",
		Link:           "http://www.link.com",
		Description:    "The description",
		Language:       "en-us",
		Copyright:      "coptyright 20032",
		ManagingEditor: "managing.editor@gmail.com (Managing Editor)",
		WebMaster:      "web.master@gmail.com (Web Master)",
		PubDate:        ComposeRssDate(time.Now()),
		LastBuildDate:  ComposeRssDate(time.Now()),
		Categories: []Category{
			{Category: "Some category"},
			{Category: "Other category", Domain: "http://domain.com"}},
		Generator: "foo",
		Docs:      DocsURL,
		Cloud: &Cloud{
			Domain:            "domain",
			Port:              80,
			Path:              "/cloud.foo",
			RegisterProcedure: "registerMe",
			Protocol:          "xml-rpc"},
		Ttl: 80,
		Image: &Image{
			Url:    "http://image.png",
			Title:  "image title",
			Link:   "http://link.com",
			Width:  80,
			Height: 120},
		Rating: "PG13",
		TextInput: &TextInput{
			Title:       "Text input title",
			Description: "Text input description",
			Name:        "The name",
			Link:        "http://www.foo.com"},
		SkipHours: &Hours{[]int{2, 12, 14}},
		SkipDays:  &Days{[]string{"Monday", "Tuesday"}},
		Items: []Item{
			{Title: "The title",
				Link:        "http://www.title.com/link",
				Description: "The item description",
				Author:      "mr.rodgers@neighborhood.com",
				Categories: []Category{
					{"foo/bar", "http://catdomain.com"}},
				Comments: "http://comment.com/",
				Enclosure: &Enclosure{
					Url:    "http://enclosure.com/foo.mp3",
					Length: 1024,
					Type:   "mpeg/audio"},
				Guid: &Guid{
					Guid: "http://guid.com", IsPermaLink: true},
				PubDate: ComposeRssDate(time.Now()),
				Source: &Source{
					Source: "thetitle",
					Url:    "http://www.foo.com"}}}}
	err = Verify(rss)
	if err != nil {
		t.Fatalf("Unable to verify maximum %v\n", err)
	}
	_, err = xml.MarshalIndent(rss, "", "    ")
	if err != nil {
		t.Fatalf("Unable to marshal minimum %v\n", err)
	}
}
