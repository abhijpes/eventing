// Copyright (c) 2017 Couchbase, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an "AS IS"
// BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing
// permissions and limitations under the License.

#ifndef UTILS_H
#define UTILS_H

#include <algorithm>
#include <curl/curl.h>
#include <libcouchbase/api3.h>
#include <libcouchbase/couchbase.h>
#include <libplatform/libplatform.h>
#include <sstream>
#include <string>
#include <unistd.h>
#include <unordered_map>
#include <v8.h>
#include <vector>

#include "comm.h"
#include "log.h"

#define DATA_SLOT 0
#define EXCEPTION_STR_SIZE 20
#define MAXPATHLEN 256

class N1QL;
class V8Worker;
class JsException;
class Transpiler;

// Struct for storing isolate data
struct Data {
  CURL *curl_handle;
  N1QL *n1ql_handle;
  V8Worker *v8worker;
  JsException *js_exception;
  Communicator *comm;
  Transpiler *transpiler;

  int fuzz_offset;
  int cron_timers_per_doc;
  lcb_t cb_instance;
  lcb_t meta_cb_instance;
};

inline Data *UnwrapData(v8::Isolate *isolate) {
  return reinterpret_cast<Data *>(isolate->GetData(DATA_SLOT));
}

template <typename T>
T *UnwrapInternalField(v8::Local<v8::Object> obj, int field_no) {
  auto field = v8::Local<v8::External>::Cast(obj->GetInternalField(field_no));
  return static_cast<T *>(field->Value());
}

int WinSprintf(char **strp, const char *fmt, ...);

v8::Local<v8::String> v8Str(v8::Isolate *isolate, const char *str);
v8::Local<v8::String> v8Str(v8::Isolate *isolate, const std::string &str);
v8::Local<v8::Name> v8Name(v8::Isolate *isolate, uint32_t key);
std::string ObjectToString(v8::Local<v8::Value> value);

std::string JSONStringify(v8::Isolate *isolate, v8::Handle<v8::Value> object);

const char *ToCString(const v8::String::Utf8Value &value);
bool ToCBool(const v8::Local<v8::Boolean> &value);

std::string ConvertToISO8601(std::string timestamp);
std::string GetTranspilerSrc();
bool isFuncReference(const v8::FunctionCallbackInfo<v8::Value> &args, int i);
std::string ExceptionString(v8::Isolate *isolate, v8::TryCatch *try_catch);

std::vector<std::string> split(const std::string &s, char delimiter);

std::string Localhost(bool isUrl);
void SetIPv6(bool is6);
bool IsIPv6();
std::string JoinHostPort(const std::string &host, const std::string &port);
std::pair<std::string, std::string> GetLocalKey();
std::string GetTimestampNow();
ParseInfo UnflattenParseInfo(std::unordered_map<std::string, std::string> &kv);

#endif
