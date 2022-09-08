# PoC HTTP redirector using Redis 6 client caching


This is a proof of concept for an HTTP redirector using Redis 6 client caching.


## Example usage:

```bash
$ http localhost:8080/                                 
HTTP/1.1 302 Found
Content-Length: 46
Content-Type: text/html; charset=utf-8
Date: Thu, 08 Sep 2022 11:30:51 GMT
Location: https://www.google.com/

<a href="https://www.google.com/">Found</a>.



```