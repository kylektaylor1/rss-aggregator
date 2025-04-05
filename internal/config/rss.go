package config

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kylektaylor1/rss-aggregator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (r *RSSFeed) unescapeHtml() {
	r.Channel.Title = html.UnescapeString(r.Channel.Title)
	r.Channel.Description = html.UnescapeString(r.Channel.Description)

	for _, item := range r.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	fmt.Println("begin fetch feed...")
	// build the request
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}

	fmt.Println("fetching request...")
	// do req with client
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	// read boy
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// marshal data
	var dataResp RSSFeed
	unmErr := xml.Unmarshal(body, &dataResp)
	if unmErr != nil {
		return nil, unmErr
	}

	dataResp.unescapeHtml()

	return &dataResp, nil
}

func ScrapeFeeds(s *State) error {
	fmt.Println("scraping...")
	nextFeed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("nextFeed: %+v\n", nextFeed)

	mfErr := s.Db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if mfErr != nil {
		fmt.Printf("error making feed fetched: %v\n", mfErr)
		return mfErr
	}
	fmt.Println("mark feed fetched")

	rssFeed, err := FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return err
	}

	fmt.Printf("Feed: %v\n", rssFeed.Channel.Title)
	for _, item := range rssFeed.Channel.Item {
		parsedTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
		if err != nil {
			fmt.Printf("error parsing time: %v\n", err)
			return err
		}
		newPost, cpErr := s.Db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			FeedID:      nextFeed.ID,
			PublishedAt: parsedTime,
		})
		if cpErr != nil {
			fmt.Printf("Error creating post: %v\n", cpErr)
		}
		fmt.Printf("Created post with title: %v\n", newPost.Title)
	}

	return err
}
