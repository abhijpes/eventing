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

#include "parse_deployment.h"

deployment_config *ParseDeployment(const char *app_code) {
  deployment_config *config = new deployment_config();

  auto app_cfg = flatbuf::cfg::GetConfig((const void *)app_code);

  auto dep_cfg = app_cfg->depCfg();
  config->metadata_bucket = dep_cfg->metadataBucket()->str();
  config->source_bucket = dep_cfg->sourceBucket()->str();

  auto buckets = dep_cfg->buckets();

  std::map<std::string, std::vector<std::string>> buckets_info;
  for (unsigned int i = 0; i < buckets->size(); i++) {
    std::vector<std::string> bucket_info;
    bucket_info.push_back(buckets->Get(i)->bucketName()->str());
    bucket_info.push_back(buckets->Get(i)->alias()->str());

    buckets_info[buckets->Get(i)->alias()->str()] = bucket_info;
  }

  config->component_configs["buckets"] = buckets_info;

  return config;
}
