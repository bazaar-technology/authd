package main



type Context struct {

	AdminKey string
	Namespace string
	Buckets map[Key]*Bucket
}

/* AllowApiKey - allow an api key across all buckets, a global api key */
func (ctx *Context) AllowApiKey(key ApiKey) (bool,error) {
	
	if !key.IsValid() {
		return false,KeyInvalid
	}

	for _,b := range ctx.Buckets {

		b.AllowApiKey(key) /* we don't care about the return */
	}
	return true,nil
}

/* RevokeApiKey - revoke an api key across all buckets on a global scale */
func (ctx *Context) RevokeApiKey(key ApiKey) (bool,error) {

	if !key.IsValid() {
		return false,KeyInvalid
	}

	for _,b := range ctx.Buckets {

		b.RevokeApiKey(key)
	}
	return true,nil
}



/* GetBucket - find a global bucket by key */
func (ctx *Context) GetBucket(key Key) *Bucket {

	if !key.IsValid() {
		return nil
	}

	if b,exists := ctx.Buckets[key]; exists {
		
		return b
	}
	return nil
}

/* AddBucket - add a new bucket to the global space, fails on existing bucket by same key */
func (ctx *Context) AddBucket(name Key) (*Bucket,error) {

	if !name.IsValid() {
		return nil,KeyInvalid
	}

	if b,exists := ctx.Buckets[name]; exists {

		return b,AlreadyPresent
	}

	b := NewBucket(name)
	
	ctx.Buckets[name] = b
	return b,nil
}

/* SetBucket - add a new bucket to the global space, if not existing add new, else return previous */
func (ctx *Context) SetBucket(name Key) (*Bucket,error) {

	if !name.IsValid() {
		return nil,KeyInvalid
	}

	if b,exists := ctx.Buckets[name]; exists {

		return b,nil
	}


	b := new(Bucket)
	b.Name = name
	b.Records = make(map[Key]Record,0)

	ctx.Buckets[name] = b
	return b,nil
}

/* DelBucket - delete bucket in the global space */
func (ctx *Context) DelBucket(name Key) (error) {

	if !name.IsValid() {
		return KeyInvalid
	}

	if _,exists := ctx.Buckets[name]; !exists {

		return NotFound
	}

	delete(ctx.Buckets,name)
	return nil
}

func NewContext() *Context {

	c := new(Context)
	c.Buckets = make(map[Key]*Bucket,0)
	return c
}

