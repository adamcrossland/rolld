# rolld
## A server that provides die-rolling capabilities for distributed clients using WebSockets

This is a little project that I put together in my spare time to learn about Web Sockets as well as to freshen up my
Go skills, which hadn't gotten much use in a little while.

This is open source, of course, and any one is free to build and spin up their own copy of the server, but for demonstration
purposes and testing, there is an up-to-date build running at http://rolld.net. If you are a frsutrated fan of FRPGs who just
can't seem to get the gang together in one place at the same time, try it out as a tool for handling die rolling fairly and
transparently.

I welcome contributors, pull requests, issues and feedback.

### CAVEAT EMPTOR:
The sample web client is specifically designed to be self-contained in one big file. I wanted it to be easily trasnportable, and
as such, it represents a number of poor choices for building good, modular web apps. Don't use it as an example for your own coding.
Go ahead and look at the code that handles the client-side Web Sockets if you want to, but for heaven's sake, keep your markup,
styling and code separate.

Additionally, the server code likely needs some reorganizing, neatening, commenting and refactoring. It's not bad, but it has grown
quickly with a strong emphasis on getting to an MVP, so I haven't yet spent much time cleaning up tech debt or commenting the code.
That's not normal for me, and it bugs me. I will get to it any day now.
