package rediscommands

// Commands is a string slice that contains all redis's command
var Commands = []string{
	"append",
	"asking",
	"auth",
	"bgrewriteaof",
	"bgsave",
	"bitcount",
	"bitfield",
	"bitop",
	"bitpos",
	"blpop",
	"brpop",
	"brpoplpush",
	"client",
	"cluster",
	"command",
	"config",
	"dbsize",
	"debug",
	"decr",
	"decrby",
	"del",
	"discard",
	"dump",
	"echo",
	"eval",
	"evalsha",
	"exec",
	"exists",
	"expire",
	"expireat",
	"flushall",
	"flushdb",
	"geoadd",
	"geodist",
	"geohash",
	"geopos",
	"georadius",
	"georadiusbymember",
	"get",
	"getbit",
	"getrange",
	"getset",
	"hdel",
	"hexists",
	"hget",
	"hgetall",
	"hincrby",
	"hincrbyfloat",
	"hkeys",
	"hlen",
	"hmget",
	"hmset",
	"hscan",
	"hset",
	"hsetnx",
	"hstrlen",
	"hvals",
	"incr",
	"incrby",
	"incrbyfloat",
	"info",
	"keys",
	"lastsave",
	"latency",
	"lindex",
	"linsert",
	"llen",
	"lpop",
	"lpush",
	"lpushx",
	"lrange",
	"lrem",
	"lset",
	"ltrim",
	"mget",
	"migrate",
	"monitor",
	"move",
	"mset",
	"msetnx",
	"multi",
	"object",
	"persist",
	"pexpire",
	"pexpireat",
	"pfadd",
	"pfcount",
	"pfdebug",
	"pfmerge",
	"pfselftest",
	"ping",
	"psetex",
	"psubscribe",
	"psync",
	"pttl",
	"publish",
	"pubsub",
	"punsubscribe",
	"quit",
	"randomkey",
	"readonly",
	"readwrite",
	"rename",
	"renamenx",
	"replconf",
	"restore",
	"restore-asking",
	"role",
	"rpop",
	"rpoplpush",
	"rpush",
	"rpushx",
	"sadd",
	"save",
	"scan",
	"scard",
	"script",
	"sdiff",
	"sdiffstore",
	"select",
	"set",
	"setbit",
	"setex",
	"setnx",
	"setrange",
	"shutdown",
	"sinter",
	"sinterstore",
	"sismember",
	"slaveof",
	"slowlog",
	"smembers",
	"smove",
	"sort",
	"spop",
	"srandmember",
	"srem",
	"sscan",
	"strlen",
	"subscribe",
	"substr",
	"sunion",
	"sunionstore",
	"sync",
	"time",
	"ttl",
	"type",
	"unlink",
	"unsubscribe",
	"unwatch",
	"wait",
	"watch",
	"zadd",
	"zcard",
	"zcount",
	"zincrby",
	"zinterstore",
	"zlexcount",
	"zrange",
	"zrangebylex",
	"zrangebyscore",
	"zrank",
	"zrem",
	"zremrangebylex",
	"zremrangebyrank",
	"zremrangebyscore",
	"zrevrange",
	"zrevrangebylex",
	"zrevrangebyscore",
	"zrevrank",
	"zscan",
	"zscore",
	"zunionstore",
}

// ForEach is a command generator
func ForEach(f func(string, ...interface{}) interface{}) []interface{} {
	var result []interface{}
	for _, command := range Commands {
		result = append(result, f(command))
	}
	return result
}
