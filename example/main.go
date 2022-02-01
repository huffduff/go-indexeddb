package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/huffduff/go-indexeddb/indexeddb"
)

type Article struct {
	Id    string   `json:"_id"`
	Title string   `json:"title"`
	Body  []string `json:"body"`
}

type Comment struct {
	Id        string `json:"_id"`
	ArticleId string `json:"articleId"`
	Author    string `json:"author"`
	Body      string `json:"comment"`
}

func (a Comment) Keys(idx string) []indexeddb.Key {
	switch idx {
	case "byAuthor":
		return []indexeddb.Key{{a.Author}}
	}
	return []indexeddb.Key{}
}

func migrate(version uint, h *indexeddb.MigrationTransaction) error {
	fmt.Printf("migrating from version %d\n", version)
	switch version {
	case 0:
		fmt.Println("initializing new database")
		_, err := h.CreateStore("article", indexeddb.StoreOptions{})
		if err != nil {
			return err
		}
		comment, err := h.CreateStore("comment", indexeddb.StoreOptions{})
		if err != nil {
			return err
		}
		return comment.CreateIndex("byAuthor", indexeddb.IndexOptions{})
	}
	panic("could not migrate database")
}

func main() {
	var path string
	flag.StringVar(&path, "path", "./", "set the root database path")
	flag.Parse()

	db, err := indexeddb.Open("blog", 1, path).Migrate(migrate)
	if err != nil {
		panic(err)
	}

	t, err := db.ReadonlyTransaction([]string{"article", "comment"}, indexeddb.Default)
	if err != nil {
		panic(err)
	}

	a := t.Store("article")

	// get all articles

	var articles []Article
	a.GetAll(indexeddb.All(), 0, &articles)
	if err != nil {
		panic(err)
	}

	fmt.Printf("found %d articles\n", len(articles))
	for _, article := range articles {
		fmt.Printf("_id: %s\n", article.Id)
		fmt.Printf("title: %s\n", article.Title)
		fmt.Printf("body: \n\t\t%s\n\n", strings.Join(article.Body, "\n\t\t"))
	}
}
