#!/bin/bash

# Copyright 2025 Duc-Hung Ho.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ruleparser-sentinez \
    -out core/waf/gen/v4-16-0 \
    -file datasets/rules/v4-16-0/REQUEST-932-APPLICATION-ATTACK-RCE.conf

ruleparser-sentinez \
    -out core/waf/gen/v4-16-0 \
    -file datasets/rules/v4-16-0/REQUEST-942-APPLICATION-ATTACK-SQLI.conf

ruleparser-sentinez \
    -out core/waf/gen \
    -file datasets/rules/setup.conf

ruleparser-sentinez \
    -out core/waf/gen \
    -file datasets/rules/default.conf

ruleparser-sentinez \
    -out core/waf/gen/v4-16-0 \
    -file datasets/rules/v4-16-0/REQUEST-901-INITIALIZATION.conf

ruleparser-sentinez \
    -out core/waf/gen \
    -file datasets/rules/REQUEST-901-INITIALIZATION.conf

ruleparser-sentinez \
    -out core/waf/gen/v4-16-0 \
    -file datasets/rules/v4-16-0/REQUEST-949-BLOCKING-EVALUATION.conf

ruleparser-sentinez \
    -out core/waf/gen \
    -file datasets/rules/REQUEST-949-BLOCKING-EVALUATION.conf

ruleparser-sentinez \
    -out core/waf/gen \
    -file datasets/rules/audit.conf