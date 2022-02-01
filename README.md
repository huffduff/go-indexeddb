# go-indexeddb

This expirement will attempt to implement IndexedDB features for Go.
It wraps [goleveldb](https://pkg.go.dev/github.com/syndtr/goleveldb@v1.0.0) similar to how Chrome's implementation utilizes LevelDB under the hood.

## Structure

| Key                             | Value                 |
| ------------------------------- | --------------------- |
| `["core"]`                      | database definition   |
| `["core", "store", <store>]`    | store spec            |
| `["core", "index", <index>]`    | index spec            |
| `["data", <store>, <id>]`       | data record           |
| `["idx", <index>, <key>]`       | index record (unique) |
| `["idx", <index>, <key>, <id>]` | index record          |

* `<store>` (string) Name of the Store
* `<index>` (string) Name of the Index
* `<id>`    (string, float, bool, nil, slice) unique identifier for a document
* `<key>` (string, float, bool, nil, slice) index key

## Roadmap

Implement the API as described by [MDN](https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API) and defined by [W3C](https://www.w3.org/TR/IndexedDB/).

### cloneable

As per the spec, objects to be stored should be supported by the [structured clone algorithm](https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API/Structured_clone_algorithm). Currently using Go's json marshalling.

### synchronicity

One obvious differentiator between Go and Javascript would be the need to make the library asynchronous.
In Go it is pretty easy to make sync libraries work async, so should be up to the user. However if there is
enough demand async wrappers could be added.

### storage limits

[Browser storage limits and eviction criteria](https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API/Browser_storage_limits_and_eviction_criteria) are beyond the scope of this project. At least initially.
