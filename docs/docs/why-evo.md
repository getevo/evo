---
layout: default
title: Why EVO?
nav_order: 6
---


The aim of EVO is to develop api/dashboard faster and more structured. So the EVO has overview over both backend and frontend concerns at the same time which let the developer concentrate more over the logic than dealing with the low level programming; In other hand EVO is based on powerful and fastest libraries which let it to achieve good performance.

* EVO is based on modular system which let developer to load only the tools that is needed.
* EVO may compile on MACOS/Linux/Windows without any restriction
* EVO may compile on arm and aarch processors along side X86
* Based on MVC


```
          Response
    +-----------------------------------------------------------------------------+
    v                                                                             |
+---+--+  Request  +----------+       +--------+       +------------+      +------+------+
| User +---------->+ FastHTTP +------>+ Router +------>+ Controller +----->+  API/View   |
+------+           +----------+       +--------+       +-----+------+      +-------------+
                                                             |
                                                             |
                                                             |
                                     +----------+      +-----+------+
                                     | Database +----->+ GORM / RDM |
                                     +----------+      +------------+

```