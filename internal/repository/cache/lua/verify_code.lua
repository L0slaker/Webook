local key = KEYS[1]
-- 用户输入的验证码
local expectedCode = ARGV[1]
local code = redis.call("get",key)
local cntKey = key..":cnt"
local cnt = tonumber(redis.call("get",cntKey))

-- 验证次数已耗尽
if cnt <= 0 then
  return -1
-- 验证码正确，消耗所有验证次数，不可再验证
elseif expectedCode == code then
    redis.call("set",cntKey,-1)
    return 0
-- 用户可能输错了，但还有验证机会，扣除一次验证次数
else
    redis.call("decr",cntKey)
    return -2
end

