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

  > authd -admin="admin-key" -addr=127.0.0.1:8080

Adding a bucket to the global space using the admin key

  PUT /api/v1/g/{bucket}
  
  > curl -XPUT -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/g/foo

Enabling a bucket in the global space - this allows a client to _see_ the bucket and its contents

  PUT /api/v1/g/{bucket}?enable=yes

  > curl -XPUT -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/g/foo?enable=yes

To disable an enabled bucket

  PUT /api/v1/g/{bucket}?disable=yes

  > curl -XPUT -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/g/foo?disable=yes

To add a key to a bucket

  PUT /api/v1/g/{bucket}/{key}

  > curl -XPUT -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/g/foo/bar

To delete a key from a bucket

  DELETE /api/v1/g/{bucket}/{key}

  > curl -XDELETE -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/g/foo/bar

To use the client interface, you need to use a valid _Api Key_ - this can be generated with

  PUT /api/v1/key

  > curl -XPUT -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/key

  74602730-7230-5d67-7d60-0400c67e8455

All _enabled_ buckets are in the global scope so _any_ valid Api Key can be used to access it

  GET /api/v1/g/{bucket}/{key}

  > curl -XGET -H "X-ApiKey:74602730-7230-5d67-7d60-0400c67e8455" http://127.0.0.1:8080/api/v1/g/foo/bar

  return of YES (200 OK) or NO (404 Not Found)

if the bucket is not enabled (is disabled by default) then checking for a key against this bucket will 
return an unauthorized response. This feature allows for staging buckets _before_ going live. 

Buckets can make use of Api Keys to make a list of authorised clients per bucket. 

  PUT /api/v1/g/{bucket}?allow={ApiKey}

  > curl -XPUT -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/g/foo?allow=74602730-7230-5d67-7d60-0400c67e8455

To revoke a Api Key on a specific bucket - 

  PUT /api/v1/g/{bucket}?revoke={ApiKey}

  > curl -XPUT -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/g/foo?revoke=74602730-7230-5d67-7d60-0400c67e8455

You can delete or revoke an api key globally with

  DELETE /api/v1/key/{api-key}

  > curl -XDELETE -H "X-AdminKey:admin-key" http://127.0.0.1:8080/api/v1/key/74602730-7230-5d67-7d60-0400c67e8455

Run _authd_ with TLS support:

  > authd -admin="admin-key" -tls -cert=/path/to/cert.pem -key=/path/to/key.pem -addr=127.0.0.1:8080