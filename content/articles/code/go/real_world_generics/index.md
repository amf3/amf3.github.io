---
title: "Go Generics: A Real World Use Case"
description: "Using Generics for Testing Pointers In Structs"
date: 2025-07-11T21:11:43-07:00
draft: false
searchHidden: false
categories: [Go, Code]
---

Until recently, I haven't had many opportunities to use Go's generics.  I ran into
a case where generics make sense.  Best of all, this isn't a contrived example.

I'm working on a project and using openAPI to generate API contracts.  One of the generated 
structs contains optional fields implemented as pointers. The only required field is Name.

```
const (
	Gzip PostPailsCompression = "gzip"
	None PostPailsCompression = "none"
)

type PostPails struct {
	Compression *PostPailsCompression `json:"compression,omitempty"`

	// MaxArchiveSize Max size (bytes) before rotating to a new archive.
	MaxArchiveSize *int `json:"max_archive_size,omitempty"`

	// Name Name of the new pail
	Name string `json:"name"`
}
```

I need to populate the struct with values when writing unit tests. But dealing with pointers in Go
test code usually results in using temporary variables.  It's not bad, but there's some visual noise.

```
gzip := PostPailsCompression("gzip")
size := 1000000
payload := PostPails{
    Name: "testpail"
    Compression: &gzip,
    MaxArchiveSize: &size,
}

```

Implementing a helper function using generics, provides a much cleaner solution.

* The temporary variables are no longer needed. 
* Test code becomes much easier to read by naming the helper function ptr.


```
func ptr[T any](v T) *T {
	return &v
}

func TestPostPails_CreatesDirectory(t *testing.T) {
	tmpStorage := t.TempDir()
	server := NewServer(tmpStorage)

	payload := PostPails{
		Name:           "testpail",
		Compression:    ptr(PostPailsCompression("gzip")),
		MaxArchiveSize: ptr(1000000),
        ... 
}
```

Let's discuss the ptr function. 

* T is a type parameter and is a placeholder for any type.
* The any constraint means T can be anything and is equivalent to interface{}.
* Inside the function, we take a value v and return its pointer.

---

Using generics avoids the temporary variable pattern and provides a means to write cleaner test code.
The benefit becomes obvious when dealing with many optional fields.

Until now, generics didn't seem to be a feature I needed.  The examples I read about didn't feel relevant.  This one clicked because 
it solved a real issue while writing unit tests.

Any thoughts or clever uses of Go generics? Drop me a line on [Bluesky](https://bsky.app/profile/af9.us).