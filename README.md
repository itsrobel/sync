---
title: sync
author: [Robel A.E. Schwarz]
sources:
  [
    https://medium.com/@abhishekranjandev/building-a-production-grade-websocket-for-notifications-with-golang-and-gin-a-detailed-guide-5b676dcfbd5a,
    https://pkg.go.dev/github.com/gorilla/websocket,
  ]
---

# Introduction

This is less of a Readme and more a design document for the development of this project

# go-watcher

# Issue Identification

I need a way to sync my notes to replace obsidian sync (go)

# Goals

I want to create an application that can be hosted on a web server as a socket server.
I would like to connect to it using my own client-selected folder(s).

The main goal for the website is to host my notes on the cloud and be able to easily download them at any time,
as well as share them using a private link system.

# Constraints

Since the web site is already being built out with gin I have to work around it

# Solution Approach

The go-watcher server will need to be assigned a folder named Alpha that it can monitor.
While the folder is being monitored, it will keep a continuous log of any changes made to
each item within the folder. When other clients connect to the go-watcher server that is
monitoring the folder Alpha, they will download the folder and then continue to monitor it
or add to the server for any future changes.

## A general list of requirements

- The client can select a folder to upload to the server and create a sync instance

  - right now the current client is a web browser instance that listens to the server
  - really the server needs to listen to the client.
  - I need to figure out how to transfer files from the client to the server

- The Server can download the folder from the client and create a "Master" copy of
- New clients that connect to the server can select which folders to then sync to
- New clients download the folder and their changes are uploaded to the server as well

## Proposed Solutions

Each of these will be Solutions on how to handle the file
differences when syncing.

1. Solution 1:
   When a file is changed for the server, re-download that file
   for each of the clients

   - Pros:
     is properly implemented in the easiest way
   - Cons:
     at scale will suck balls and requires a lot of network usage for each iteration

Even with the file transfer system, I need to atleast keep track of the movement of the files/
what their names are so I do not re download the entire file system each time a file change is made

2. Solution 2:
   When a file is changed for the server, for each client send out the difference

   - Pros
     is probably the best and maybe the most "fun" to implement
   - Cons
     requirements are much higher

   What would I need for this solution?

   I need a way for each of the clients to have a master state of machine
   or each of the files and their current "version"?

   How do I file version?

# Implementation Plan

The first thing that Is required effectively for both solutions is file versioning and send out what changes where made and to which files they were made to.

- There can be two ways to version a directory versioning and a individual file versioning.

  - the directory versioning would update very 6hrs -> 12hrs and would be a master snap shot of all the file versions
    uploaded to the server
    - the directory versioning would also be able to handle file movement/renaming/and deletion
    - this would handle the problem of what happens when a client has not connected to the server for a few days
      the client would iterate through the changes made in the ledger of the client and server to point to the
      file changes.
    - each of the file versions would have to have a timestamp ID and what changes were made to the file

## Data Structures

To keep track of these versioning, I need the project to be hooked up to a database.
In terms of data structures and what the database might look like.

```json directory
{
  "timestamp": "2022-01-01 00:00:00",
  "files-changed": [
    <!--for-each file changed index 0 is the old version and index 1 is the new-->
    "file-name": ["file{name}{timestamp}", "version-2"]
  ],
  "files-deleted": [
    "file-name1"
    "file-name2"
  ]
}
```

```json file
{
  "id": "<the uuid of the file>",
  "file-name": "<name of the file given from the user",
  "contents": "<contents of the file>",
  "location": "<location of the file in the directory",
  "changes": ["file-change1", "file-change..."]
}
```

```json file-change-{timestamp}
{
  it might be better to just use the time stamp as part of the version name
  <!--"timestamp": "2022-01-01 00:00:00"-->
  "device":"<device of where the change was made from>",
  "file-diff": "<content changes of the file"
}
```

> I am probably going to use mongodb since Ion really know anything else that well and I don't
> feel like adding to the learning curve of this project.

## Todo Items

- [ ] attach the directory watch system to the socket server
- [ ] create a socket message that downloads the file from client to server and vise versa
- [ ] send the file changes over json
- [ ] attach go to mongodb

# Risks and Challenges

- Identify any potential risks, challenges, or obstacles that may arise.

# Lessons learned/Lessons to Learn

- Reflect on lessons learned during the implementation process.

# Conclusion

- Summarize the key points and reiterate the importance of the solution-oriented approach.
