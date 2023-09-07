-- 发送到的 key，也就是 code:业务:手机号
local key = KEYS[1]
-- 验证次数，即还可以验证 3 次，code:业务:手机号:cnt
local cntKey = key..":cnt"
-- 验证码
local val = ARGV[1]

-- 过期时间 600 秒，十分钟；转成number
local ttl = tonumber(redis.call("ttl",key))
-- -1是 key 存在，但没有过期时间
if ttl == -1 then
    -- 有人误操作，导致 key 冲突
    return -2
-- -2是 key 不存在，ttl < 540 是发了一个验证码，已经超过了一分钟
elseif ttl == -2 or ttl < 540 then
    -- 后续如果验证码有不同的过期时间，此处需要优化
    redis.call("set",key,val)
    redis.call("expire",key,600)
    redis.call("set",cntKey,3)
    redis.call("expire",cntKey,600)
    return 0
else
    -- 已经发送了一个验证码，但还不到一分钟
    return -1
end