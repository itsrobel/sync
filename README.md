---
title: sync
author: [Robel A.E. Schwarz]
sources:
  [
    https://medium.com/@abhishekranjandev/building-a-production-grade-websocket-for-notifications-with-golang-and-gin-a-detailed-guide-5b676dcfbd5a,
    https://pkg.go.dev/github.com/gorilla/websocket,
  ]
---

## Introduction

This is less of a Readme and more a design document for the development of this project

## Issue Identification

I need a way to sync my notes to replace obsidian sync (go)

## Goals

I want to create an application that can be hosted on a web server as a socket server.
I would like to connect to it using my own client-selected folder(s).

The main goal for the website is to host my notes on the cloud and be able to
easily download them at any time,
as well as share them using a private link system.

## Constraints

Since the web site is already being built out with gin I have to work around it

## Solution Approach

The go-watcher server will need to be assigned a folder named Alpha that
it can monitor. While the folder is being monitored, it will keep a
continuous log of any changes made to each item within the folder. When
other clients connect to the go-watcher server that is monitoring the folder
Alpha, they will download the folder and then continue to monitor it or add to
the server for any future changes.

## A general list of requirements

- The client can select a folder to upload to the server and create a sync instance

  - right now the current client is a web browser instance that listens to the server
  - really the server needs to listen to the client.
  - I need to figure out how to transfer files from the client to the server

- The Server can download the folder from the client and create a "Master" copy of
- New clients that connect to the server can select which folders to then sync to
- New clients download the folder and their changes are uploaded to the server
  as well

## Proposed Solutions

Each of these will be Solutions on how to handle the file
differences when syncing.

### Solution 1

When a file is changed for the server, re-download that file
for each of the clients

- Pros:
  is properly implemented in the easiest way
- Cons:
  at scale will suck balls and requires a lot of network usage for each iteration

Even with the file transfer system, I need to at least keep track of the
movement of the files/ what their names are so I do not re download the entire
file system each time a file change is made

### Solution 2

When a file is changed for the server, for each client send out the difference

- Pros
  is probably the best and maybe the most "fun" to implement
- Cons
  requirements are much higher

What would I need for this solution?

I need a way for each of the clients to have a master state of machine
or each of the files and their current "version"?

How do I file version?

## Implementation Plan

The first thing that Is required effectively for both solutions is file
versioning and send out what changes where made and to which files they were
made to.

- There can be two ways to version a directory versioning and a individual file
  versioning.

  - the directory versioning would update very 6hrs -> 12hrs and would be a
    master snap shot of all the file versions
    uploaded to the server
    - the directory versioning would also be able to handle file
      movement/renaming/and deletion
    - this would handle the problem of what happens when a client
      has not connected to the server for a few days
      the client would iterate through the changes made in the ledger of the
      client and server to point to the file changes.
    - each of the file versions would have to have a timestamp ID and what
      changes were made to the file

## Data Structures

To keep track of these versioning, I need the project to be hooked up to a database.
In terms of data structures and what the database might look like.

How do I handle non markdown files?

```json file
{
  "id": "<the uuid of the file>",
  "file-name": "<name of the file given from the user",
  "location": "<location of the file in the directory",
  "contents": "<contents of the file>",
  "status": "<active/deleted>"
}
```

Any changes from the user from the client would be made to the
would point to the id of the file by searching with the location string

If the file is renamed on the server and tries to be synced, if the
file id matches then the location of the latest timestamp is taken

The ID system can be used for conflict resolution

```json file-change-{timestamp}
{
  "device": "<device of where the change was made from>",
  "id": "<the id of the file>",
  "location": "<location of the file",
  "file-diff": "<content changes of the file>",
  "status": "<deleted/moved/edited>"
}
```

If the file status is deleted for more than 30 days or some shit I can remove
the contents and the ID from the sql table

## Todo Items

- [x] attach the directory watch to the client
- [x] move the project into using gRPC since it fixes most of my problems

  - right now I have bi directional streaming of files but the files append
    and do "sync"
  - I need to change how the files are opened.
  - I also now have to re implement the database.
  - I do not need web sockets since the bi directional streaming can handle the changes

- [x] create a socket message that downloads the file from client to server
      and vise versa
- [x] figure out how to do file difs
- [x]create the data structures for database and how to query them
- [x] attach go to database

  - [x] attach go client to sqlite
  - [x] attch go server to mongodb

  - [ ] connect the updated client to the server
  - [ ] reconstruct the file data based on what is in servers mongodb tables
  - [x] create a tooling to list the changes and the versions made to a file
  - [x] make a on start call that populates database with the un-tracked files

- [x] dockerize
- [x] make buf.yamls
- [x] maybe use connectrpc
- [ ] merge the client and server data calls with connect rpc to the watcher

## I need to focus on finishing the project into a working demo duck everything else

- [x] the latest version of the file will just be the file itself
- [ ] I need a function to get the latest version of a file
- [x] when the client starts check all files in the target directory

  - a new version of each file is created

- [x] try server connection on client start
- [x] upload files if connection is made
      have the connection request be on conditional
- [x] update send the file update to server when client is connected
- [x] make the ping/pong or the server to client connection to check live

- [x] setup bi directional streaming for session handling

  - [x] the default behavior for the control stream is to constantly re check if the server is live
        since part of that logic is sending down updated files we do not have to make another controller for the file downloads
        and instead can send them as a return stream response
        this then makes almost all my rpc requests a bi directional stream

        client connects to server -> server sends down updated files from mongodb

        since the files are not created deleted or moved directly from the server


        the server data structures can be simpler. We can assume.
          - files do not neeed their own struture from version's since we are
          just appending or editing information like a log




        client uploads files -> server saves it

  - [x] I need the server to to be able to upload files to the server on start
  - [x] each client is given a session id, but I think giving them a self assigned name would be
        better, to track what files, they don't have
        based on when they lasted connected via timestamps

- [ ] make a rpc call on the web app that reads the latest list of files
- [ ] Re structure the project so that there is only one copy of each file in "files" that is
      updated on client writes to the latest version

- [x] restructure the application for max code re use among each client and server
  - [x] basically put each exe in cmd and everything else in internal
- [ ] change the font of the editor js thing to something that doesn't look ugly

### Bonus

- [ ] The file_versions need to be diffs in a order that can than
      reconstructed on file_ID request
- [ ] on the server separate out the file and file versions into a separate, for file building based on diffs

## Risks and Challenges

- Identify any potential risks, challenges, or obstacles that may arise.

## Lessons learned/Lessons to Learn

- Reflect on lessons learned during the implementation process.

## Conclusion

- Summarize the key points and reiterate the importance of the
  solution-oriented approach.

I need a interval period of file changes like a memory queue that keeps track
of the things that happen to the file and the current state
