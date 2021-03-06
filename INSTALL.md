# HOWTO Install Rolld
## And build it and configure and stuff, too...
1. Clone the Rolld repo: `git@github.com:adamcrossland/rolld.git`

2. Build it:
    
    `cd server; go build`

3. Copy the executable, `server` and the
client file `rolld-client.html` to wherever they will live on your server.

4. Set up these environment variables:

    `ROLLD_SERVER_ADDRESS` -- this will usually just be the port number on which you wish to serve Rolld. A typical value would be `":443"`

    `ROLLD_DATABASE_FILE` -- the name of the file in which your Rolld database will be stored. `"rolld.db"` would be a typical choice.

    `ROLLD_SERVER_CERTPATH` -- the path to your SSL certificate file. If you are using Let's Encrypt's CertBot, it would be something like `"/etc/letsencrypt/live/servername/fullchain.pem"`

    `ROLLD_SERVER_KEYPATH` -- the path to your SSL key file. Again, CertBot would give you something like `"/etc/letsencrypt/live/servername/privkey.pem"`

5. Use your OS-specific means to get the `server` executable running and staying running.

## CAVEAT EMPTOR
At this time (Oct 18 2018), the server process expects to be able to read from the UNIX-specific file /dev/urandom, so it will not run on
OSs. That do not have this file. If you want to build and run Rolld on your Windows machine, consider using WSL. This dependency
has an Issue assocuiated with it, and it will be removed in the relatively near future.

## At this point, you should have a running instance of Rolld.
