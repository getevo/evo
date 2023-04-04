package http

// Get execute a http GET request
func Get(url string, v ...interface{}) (*Resp, error) {
	return std.Get(url, v...)
}

// Post execute a http POST request
func Post(url string, v ...interface{}) (*Resp, error) {
	return std.Post(url, v...)
}

// Put execute a http PUT request
func Put(url string, v ...interface{}) (*Resp, error) {
	return std.Put(url, v...)
}

// Head execute a http HEAD request
func Head(url string, v ...interface{}) (*Resp, error) {
	return std.Head(url, v...)
}

// Options execute a http OPTIONS request
func Options(url string, v ...interface{}) (*Resp, error) {
	return std.Options(url, v...)
}

// Delete execute a http DELETE request
func Delete(url string, v ...interface{}) (*Resp, error) {
	return std.Delete(url, v...)
}

// Patch execute a http PATCH request
func Patch(url string, v ...interface{}) (*Resp, error) {
	return std.Patch(url, v...)
}

// Do execute request.
func Do(method, url string, v ...interface{}) (*Resp, error) {
	return std.Do(method, url, v...)
}
