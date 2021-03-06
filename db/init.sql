-- name: drop-tables
DROP TABLE IF EXISTS votes CASCADE;
DROP TABLE IF EXISTS items CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;
DROP TABLE IF EXISTS instances CASCADE;
DROP TABLE IF EXISTS objects CASCADE;
-- DROP TABLE IF EXISTS activities CASCADE;
-- DROP TABLE IF EXISTS actors CASCADE;
DROP TABLE IF EXISTS access CASCADE;
DROP TABLE IF EXISTS authorize CASCADE;
DROP TABLE IF EXISTS refresh CASCADE;
DROP TABLE IF EXISTS client CASCADE;
DROP TYPE  IF EXISTS "types";

-- name: truncate-tables
TRUNCATE votes RESTART IDENTITY CASCADE;
TRUNCATE accounts RESTART IDENTITY CASCADE;
TRUNCATE items RESTART IDENTITY CASCADE;
TRUNCATE instances RESTART IDENTITY CASCADE;
TRUNCATE objects RESTART IDENTITY CASCADE;
-- TRUNCATE activities RESTART IDENTITY CASCADE;
-- TRUNCATE actors RESTART IDENTITY CASCADE;
TRUNCATE access RESTART IDENTITY CASCADE;
TRUNCATE authorize RESTART IDENTITY CASCADE;
TRUNCATE refresh RESTART IDENTITY CASCADE;
TRUNCATE client RESTART IDENTITY CASCADE;

-- name: create-accounts
create table accounts (
  id serial constraint accounts_pk primary key,
  key char(32) unique,
  handle varchar,
  email varchar unique,
  score bigint default 0,
  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp,
  metadata jsonb default '{}',
  flags bit(8) default 0::bit(8)
);

-- name: create-items
create table items (
  id serial constraint items_pk primary key,
  key char(32) unique,
  mime_type varchar default NULL,
  title varchar default NULL,
  data text default NULL,
  score bigint default 0,
  path ltree default NULL,
  submitted_by int references accounts(id),
  submitted_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp,
  metadata jsonb default '{}',
  flags bit(8) default 0::bit(8)
);

-- name: create-votes
create table votes (
  id serial constraint votes_pk primary key,
  submitted_by int references accounts(id),
  submitted_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp,
  item_id  int references items(id),
  weight int,
  flags bit(8) default 0::bit(8),
  constraint unique_vote_submitted_item unique (submitted_by, item_id)
);

-- name: create-instances
create table instances
(
  id serial constraint instances_pk primary key,
  name varchar not null,
  description text default NULL,
  url varchar not null constraint instances_url_key unique,
  inbox varchar,
  metadata jsonb default '{}',
  flags bit(8) default 0::bit(8)
);


-- name: create-activitypub-types-enum
CREATE TYPE "types" AS ENUM (
  'Object',
  'Link',
  'Activity',
  'IntransitiveActivity',
  'Actor',
  'Collection',
  'OrderedCollection',
  'CollectionPage',
  'OrderedCollectionPage',
  'Article',
  'Audio',
  'Document',
  'Event',
  'Image',
  'Note',
  'Page',
  'Place',
  'Profile',
  'Relationship',
  'Tombstone',
  'Video',
  'Mention',
  'Application',
  'Group',
  'Organization',
  'Person',
  'Service',
  'Accept',
  'Add',
  'Announce',
  'Arrive',
  'Block',
  'Create',
  'Delete',
  'Dislike',
  'Flag',
  'Follow',
  'Ignore',
  'Invite',
  'Join',
  'Leave',
  'Like',
  'Listen',
  'Move',
  'Offer',
  'Question',
  'Reject',
  'Read',
  'Remove',
  'TentativeReject',
  'TentativeAccept',
  'Travel',
  'Undo',
  'Update',
  'View'
  );

-- name: create-activitypub-objects
create table objects
(
  "id"  serial not null constraint objects_pkey primary key,
  "key" char(32) constraint objects_key_key unique,
  "iri" varchar constraint objects_iri_key unique,
  "type" types,
  "raw" jsonb
);

-- name: create-activitypub-actors
create table actors (
  "id" serial not null constraint actors_pkey primary key,
  "key" char(32) constraint actors_key_key unique,
  "account_id" int default NULL, -- the account for this actor
  "type" varchar, -- maybe enum
  "pub_id" varchar, -- the activitypub Object ID (APIURL/self/following/{key})
  "url" varchar, -- frontend reachable url
  "name" varchar,
  "preferred_username" varchar,
  "published" timestamp default CURRENT_TIMESTAMP,
  "updated" timestamp default CURRENT_TIMESTAMP,
  -- "inbox_id" int,
  "inbox" varchar,
  -- "outbox_id" int,
  "outbox" varchar,
  -- "liked_id" int,
  "liked" varchar,
  -- "followed_id" int,
  "followed" varchar,
  -- "following_id" int,
  "following" varchar
);

-- this is used to store the Activtities we're receiving in outboxes and inboxes
-- name: create-activitypub-activities
create table activities (
  "id" serial not null constraint activities_pkey primary key,
  "key" char(32) constraint activities_key_key unique,
  "pub_id" varchar, -- the activitypub Object ID
  "actor_id" int default NULL, -- the actor id, if this is a local activity
  "account_id" int default NULL, -- the account id, if this is a local actor
  "actor" varchar, -- the IRI of local or remote actor
  "object_id" int default NULL, -- the object id if it's a local object
  "item_id" int default NULL, -- the item id if it's a local object
  "object" varchar, -- the IRI of the local or remote object
  "published" timestamp default CURRENT_TIMESTAMP,
  "audience" jsonb -- the [to, cc, bto, bcc fields]
);

-- this is used to store Note/Article objects that correspond to elements in the items table
-- name: __create-activitypub-objects
create table objects (
  "id" serial not null constraint objects_pkey primary key,
  "key" char(32) constraint objects_key_key unique,
  "pub_id" varchar, -- the activitypub Object ID
  "type" varchar, -- maybe enum
  "url" varchar,
  "name" varchar,
  "published" timestamp default CURRENT_TIMESTAMP,
  "updated" timestamp default CURRENT_TIMESTAMP
);

-- oauth for osin
-- name: create-oauth-storage
CREATE TABLE IF NOT EXISTS client (
   id varchar NOT NULL PRIMARY KEY,
   secret varchar NOT NULL,
   extra jsonb DEFAULT NULL,
   redirect_uri varchar NOT NULL
 );
CREATE TABLE IF NOT EXISTS authorize (
  client varchar NOT NULL,
  code varchar NOT NULL PRIMARY KEY,
  expires_in int NOT NULL,
  scope varchar NOT NULL,
  redirect_uri varchar NOT NULL,
  state varchar NOT NULL,
  extra jsonb DEFAULT NULL,
  created_at timestamp with time zone NOT NULL
);
CREATE TABLE IF NOT EXISTS access (
  client varchar NOT NULL,
  authorize varchar NOT NULL,
  previous varchar NOT NULL,
  access_token varchar NOT NULL PRIMARY KEY,
  refresh_token varchar NOT NULL,
  expires_in int NOT NULL,
  scope varchar NOT NULL,
  redirect_uri varchar NOT NULL,
  extra jsonb DEFAULT NULL,
  created_at timestamp with time zone NOT NULL
);
CREATE TABLE IF NOT EXISTS refresh (
  token varchar NOT NULL PRIMARY KEY,
  access varchar NOT NULL
);
