-- CREATE DATABASE spacexlaunchbotsite
-- \c spacexlaunchbotsite

CREATE TYPE notification AS ENUM ('all', 'schedule', 'launch');

CREATE TABLE guild (
    guild_id text primary key,
    guild_name text
);

CREATE TABLE channel (
    channel_id text primary key,
    channel_name text,
    guild_id text references guild
);

CREATE TABLE subscribed_channels (
    channel_id text primary key references channel,
    notification_type notification,
    launch_mentions text
);