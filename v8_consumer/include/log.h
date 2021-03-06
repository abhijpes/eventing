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

#ifndef LOG_H
#define LOG_H

#include <atomic>
#include <chrono>
#include <ctime>
#include <iomanip>
#include <iostream>
#include <mutex>
#include <sstream>
#include <string>

enum LogLevel { logSilent, logError, logInfo, logWarning, logDebug, logTrace };
extern std::string appName;
extern LogLevel desiredLogLevel;
extern std::string workerID;
extern bool noRedact;

inline std::string NowTime();

extern void setAppName(std::string appName);
extern void setLogLevel(LogLevel level);
extern void setWorkerID(std::string ID);

inline LogLevel LevelFromString(const std::string &level) {
  if (level == "SILENT")
    return logSilent;
  if (level == "INFO")
    return logInfo;
  if (level == "ERROR")
    return logError;
  if (level == "WARNING")
    return logWarning;
  if (level == "DEBUG")
    return logDebug;
  if (level == "TRACE")
    return logTrace;

  return logInfo;
}

class AtomicLog {
public:
  AtomicLog() {
    while (spin_lock.test_and_set(std::memory_order_acquire)) {
    }
  }

  std::ostream &Cout() { return std::cout; }

  ~AtomicLog() { spin_lock.clear(std::memory_order_release); }

  static std::atomic_flag spin_lock;
};

#define LOG(level)                                                             \
  if (level > desiredLogLevel)                                                 \
    ;                                                                          \
  else                                                                         \
    AtomicLog().Cout()
#endif

#define RU(msg) (noRedact ? "" : "<ud>") << msg << (noRedact ? "" : "</ud>")
#define RM(msg) msg
#define RS(msg) msg
