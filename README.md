WriteToN
-------

writeton package is workaround io.CopyN and io.LimitedReader limitation

Which ever some one uses io.CopyN or wraps io.Reader by LimitedReader,
there is no simple way to suggest io.Copy that there is more efficient
way to do the copy than by Read/Write cycle.

This is because LimitedReader doesn't provide way to do such suggest: it
doesn't provide WriteTo methods and do not test wrapped Reader for any
extension method.

net.TCPConn.ReadFrom handles case when LimitedReader wraps *os.File. But
if wrapped io.Reader is not *os.File then it "removes" ReadFrom method,
and there is no way to get it back. Even if we just do second wrap
(ie io.CopyN(socket, io.LimitReader(file, n), m), then net.TCPConn will
fail to recognize that it could use sendfile.

Look at the test in net/http in the changeset: simple muxer (for example,
which duplicates file content) can not be passed to http.ServeContent
and still use sendfile without improving io.LimitedReader.

Within realword proxy-caching file server, there could be really tangled
mix of rate limiter, proxy-cacher (which takes file from backend storage)
and file remuxer (video clipping/streaming, for example). And it will be
not trivial to untangle it to safely use sendfile.

Natural way to solve thing should be extending io.LimitedReader, but
proposal is not accepted
https://go-review.googlesource.com/#/c/7893/
https://groups.google.com/forum/#!topic/golang-dev/D7tO5HZTs60
http://pastebin.com/zCVrzVx6

So, this package provides alternative way to unpack all io.LimitedReader
and use all WriteToN methods recursively.

You should just from io.Writer with writeton.Writer:

````go
     w := some_w.(io.Writer)
     wr := writeton.Writer{w}
     io.Copy(wr, some_r.(io.Reader))
     io.CopyN(wr, some_r.(io.Reader), n)
````

Or just use writeton.Copy and writeton.CopyN:
````go
     w := some_w.(io.Writer)
     writeton.Copy(wr, some_r.(io.Reader))
     writeton.CopyN(wr, some_r.(io.Reader), n)
````

Or wrap http.ResponseWriter if you wish to pass complex io.ReadSeeker to
http.ServeContent:

````go
     func handler(rw http.ResponseWriter, rq *http.Request) {
         var r io.ReadSeeker
         var m time.Time
         r, m = SomeComplexFileObject(rq)
         smartrw := writeton.NewResponseWriter(rw)
         http.ServeContent(smartrw, req, "", m, r)
     }
````

Copyright (c) 2015, Sokolov Yura aka funny_falcon

The code is in public domain.
I will be a bit happier if you mention me with good word.
