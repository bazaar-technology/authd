authd
=====

Authorise Users/Agents from a http login before hitting a 
local user/password database.

A key is stored inmemory and can be populated and checked from
a http interface. Useful as a component between your login application and
a user/password database, esp. as you do not need to involve a password in
the initial check.  
