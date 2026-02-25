---
Title: SUMMARY
Ticket: SURF-20260225-R2
Status: active
Topics:
  - go
  - chromium
  - native-messaging
DocType: reference
Intent: working
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Generated comparison artifact"
LastUpdated: 2026-02-25T18:51:44-05:00
WhatFor: "Captured investigation output artifact"
WhenToUse: "Use when reviewing output-format investigation evidence"
---

# Node vs Go Output Comparison

- Socket: `/home/manuel/snap/chromium/common/surf-cli/surf.sock`
- Generated: 2026-02-25T23:48:29.291Z

## Setup

- tab.new stdout: `Created tab 441388280: https://example.com`

## Cases

### tab-list
- Node: ok=true status=0 shape=array(3)<object{id,title,url,active,windowId}>
- Go: ok=true status=0 shape=array(1)<object{content,data,data_count,data_kind,error,id,result,status}>
### page-read
- Node: ok=true status=0 shape=string
- Go: ok=true status=0 shape=array(1)<object{content,data,data_count,data_kind,error,id,result,status}>
### page-text
- Node: ok=true status=0 shape=string
- Go: ok=true status=0 shape=array(1)<object{content,data,data_count,data_kind,error,id,result,status}>
### page-state
- Node: ok=true status=0 shape=object{id,focusedElement,hasDatePicker,hasDropdown,hasModal,hasOverlay,title,url}
- Go: ok=true status=0 shape=array(1)<object{content,data,data_count,data_kind,error,id,result,status}>
### network-list
- Node: ok=true status=0 shape=string
- Go: ok=true status=0 shape=array(1)<object{content,data,data_count,data_kind,error,id,result,status}>
### console-read
- Node: ok=true status=0 shape=string
- Go: ok=true status=0 shape=array(1)<object{content,data,data_count,data_kind,error,id,result,status}>
### navigate
- Node: ok=true status=0 shape=string
- Go: ok=true status=0 shape=array(1)<object{content,data,data_count,data_kind,error,id,result,status}>
