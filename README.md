# go-indexeddb

This expirement will attempt to implement IndexedDB features for Go.
It wraps [goleveldb](https://pkg.go.dev/github.com/syndtr/goleveldb@v1.0.0) similar to how Chrome's implementation utilizes LevelDB under the hood.

## Concept

Each record key is serialized from the store name and assigned key.

`["core", <store name>]`
<!-- `["core", <store name>, "idx", <index name>]` -->
`["data", <store name>, <key>]`
`["idx", <store name>, <index name>, <key>]`

## Roadmap

Implement the API as described by [MDN](https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API) and defined by [W3C](https://www.w3.org/TR/IndexedDB/).

### cloneable

As per the spec, objects to be stored should be supported by the [structured clone algorithm](https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API/Structured_clone_algorithm). Currently we're using Go's json marshalling.

### synchronicity

One obvious differentiator between Go and Javascript would be the need to make the library asynchronous.
In go it is pretty easy to make sync libraries work async, so should be up to the user. If there is
enough demand however, we could always implement async wrappers.

### storage limits

[Browser storage limits and eviction criteria](https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API/Browser_storage_limits_and_eviction_criteria) are beyond the scope of this project. At least initially.
