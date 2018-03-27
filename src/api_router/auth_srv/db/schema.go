package db

var (
	accountSchema = `CREATE TABLE IF NOT EXISTS accounts (
licensekey varchar(127) primary key,
username varchar(255),
pubkey varchar(2048),
created integer,
unique (username));`
)
