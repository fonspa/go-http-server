# Cookies

A piece of data that a server sends to a client. The client stores the cookie and sends it back to the server on subsequent requests.

Cookies can store arbitrary data:
- tracking information
- JWT
- Items in a shopping cart
- etc.

Cookies work through HTTP headers. They are sent by the server in the `Set-Cookie` header.

Cookies are more popular for browser-based applications because browsers automatically send cookies they have back to the server in the `Cookie` header.

=> A good use-case for cookies is to serve as a more *strict and secure transport layer for JWTs*. You can ensure that 3rd party JavaScript being executed on the website can't access any cookies. That's a lot better than storing JWTs in the browser's local storage, where it's easily accessible by any JavaScript running on the page.
