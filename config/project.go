package config

// 用户注册密码加盐
const UserSignupSalt = "*#890"

// const FileLocalStorePath = "/tmp/"
const FileLocalStorePath = "C:/_test_object_storage/"
const FileCephStorePath = "/ceph"
const FileOSSStorePath = "oss/"

// TODO - NOTICE Go语言时间格式规则
const SysTimeform = "2006-01-02 15:04:05"
const SysTimeformShort = "2006-01-02"

const RdsCacheKeyPrefix = "object_storage:"
