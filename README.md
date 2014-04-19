authd
=====

**Currently under develpment, do not use for production*** 

Authorise Users/Agents from a http login before hitting a 
local user/password database.

A key is stored inmemory and can be populated and checked from
a http interface. Useful as a component between your login application and
a user/password database, esp. as you do not need to involve a password in
the initial check.  


Running Authd
-------------

Run _authd_ in a secured environment over http

  > authd -admin="admin-key" -addr=loopback:8888

Adding a bucket to the global space:

  /api/v1/add/bucket-name/

  > curl -H "X-AdminKey:admin-key" http://loopback:8888/api/v1/add/foo/

Setting a bucket to the global space:

  /api/v1/set/bucket-name/
 
  > curl -H "X-AdminKey:admin-key" http://loopback:8888/api/v1/set/foo/

Deleting a bucket from the global space:

  /api/v1/del/bucket-name/

  > curl -H "X-AdminKey:admin-key" http://loopback:8888/api/v1/del/foo/

Enable a bucket - this makes the bucket _checkable_ in the global space:

  /api/v1/enable/bucket-name/

  > curl -H "X-AdminKey:admin-key" http://loopback:8888/api/v1/enable/foo/

To disable a bucket:

  /api/v1/disable/bucket-name/

  > curl -H "X-AdminKey:admin-key" http://loopback:8888/api/v1/disable/foo/

To add a key to a bucket:

  /api/v1/add/bucket-name/key-name/

  > curl -H "X-AdminKey:admin-key" http://loopback:8888/api/v1/add/foo/bar/

To set a key:

  /api/v1/set/bucket-name/key-name/

  > curl -H "X-AdminKey:admin-key" http://loopback:8888/api/v1/set/foo/bar/

To delete a key:

  /api/v1/del/bucket-name/key-name/

  > curl -H "X-AdminKey:admin-key" http://loopback:8888/api/v1/del/foo/bar/

To check a key in a bucket:
  
  /api/v1/check/bucket-name/key-name/

  > curl http://loopback:8888/api/v1/check/foo/bar/

if the bucket is not enabled (is disabled by default) then checking for a key against this bucket will 
return an unauthorized response. This feature allows for staging buckets _before_ going live. 




Run _authd_ with TLS support:

  > authd -admin="admin-key" -tls -cert=/path/to/cert.pem -key=/path/to/key.pem -addr=loopback:8888