create database entry;

create user entry@'%' identified by 'password';

grant select, insert, update(status, ended_at, updated_at) on entry.sessions to entry@'%';
grant select, insert on entry.commands to entry@'%';
flush privileges;
