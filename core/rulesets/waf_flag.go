// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package corers

type Flag uint64

const (
	// ====== REQUEST PHASE ======
	AllCRS Flag = 1 << iota

	ReqCommonExceptions    // REQUEST-905-COMMON-EXCEPTIONS.conf
	ReqMethodEnforcement   // REQUEST-911-METHOD-ENFORCEMENT.conf
	ReqScannerDetection    // REQUEST-913-SCANNER-DETECTION.conf
	ReqProtocolEnforcement // REQUEST-920-PROTOCOL-ENFORCEMENT.conf
	ReqProtocolAttack      // REQUEST-921-PROTOCOL-ATTACK.conf
	ReqMultipartAttack     // REQUEST-922-MULTIPART-ATTACK.conf
	ReqAppAttackLFI        // REQUEST-930-APPLICATION-ATTACK-LFI.conf
	ReqAppAttackRFI        // REQUEST-931-APPLICATION-ATTACK-RFI.conf
	ReqAppAttackRCE        // REQUEST-932-APPLICATION-ATTACK-RCE.conf
	ReqAppAttackPHP        // REQUEST-933-APPLICATION-ATTACK-PHP.conf
	ReqAppAttackGeneric    // REQUEST-934-APPLICATION-ATTACK-GENERIC.conf
	ReqAppAttackXSS        // REQUEST-941-APPLICATION-ATTACK-XSS.conf
	ReqAppAttackSQLI       // REQUEST-942-APPLICATION-ATTACK-SQLI.conf

	// REQUEST-943-APPLICATION-ATTACK-SESSION-FIXATION.conf
	ReqAppAttackSessionFixation

	ReqAppAttackJava // REQUEST-944-APPLICATION-ATTACK-JAVA.conf

	// ====== RESPONSE PHASE ======
	RespDataLeakages         // RESPONSE-950-DATA-LEAKAGES.conf
	RespDataLeakagesSQL      // RESPONSE-951-DATA-LEAKAGES-SQL.conf
	RespDataLeakagesJava     // RESPONSE-952-DATA-LEAKAGES-JAVA.conf
	RespDataLeakagesPHP      // RESPONSE-953-DATA-LEAKAGES-PHP.conf
	RespDataLeakagesIIS      // RESPONSE-954-DATA-LEAKAGES-IIS.conf
	RespWebShells            // RESPONSE-955-WEB-SHELLS.conf
	RespDataLeakagesRuby     // RESPONSE-956-DATA-LEAKAGES-RUBY.conf
	RespBlockingEvalResponse // RESPONSE-959-BLOCKING-EVALUATION.conf
	RespCorrelation          // RESPONSE-980-CORRELATION.conf
)
