// Code generated by goyacc -o parser.y.go -v  parser.y. DO NOT EDIT.

//line parser.y:32
package smi

import __yyfmt__ "fmt"

//line parser.y:32

import (
	"strings"
)

//line parser.y:40
type yySymType struct {
	yys            int
	token          Token
	augments       string
	description    string
	imports        []Import
	importIDs      []string
	indexes        []string
	modules        []*parseModule
	object         *parseObject
	objects        []*parseObject
	objectMap      map[string]*parseObject
	orphans        []*parseObject
	status         Status
	table          bool
	val            string
	subidentifiers []string
}

const ACCESS = 57346
const AGENT_CAPABILITIES = 57347
const APPLICATION = 57348
const AUGMENTS = 57349
const BEGIN = 57350
const BIN_STRING = 57351
const BITS = 57352
const CHOICE = 57353
const COLON_COLON_EQUAL = 57354
const COMMENT = 57355
const CONTACT_INFO = 57356
const CREATION_REQUIRES = 57357
const COUNTER = 57358
const COUNTER32 = 57359
const COUNTER64 = 57360
const DEFINITIONS = 57361
const DEFVAL = 57362
const DESCRIPTION = 57363
const DISPLAY_HINT = 57364
const DOT_DOT = 57365
const END = 57366
const ENTERPRISE = 57367
const EXPORTS = 57368
const EXTENDS = 57369
const FROM = 57370
const GROUP = 57371
const GAUGE = 57372
const GAUGE32 = 57373
const HEX_STRING = 57374
const IDENTIFIER = 57375
const IMPLICIT = 57376
const IMPLIED = 57377
const IMPORTS = 57378
const INCLUDES = 57379
const INDEX = 57380
const INSTALL_ERRORS = 57381
const INTEGER = 57382
const INTEGER32 = 57383
const INTEGER64 = 57384
const IPADDRESS = 57385
const LAST_UPDATED = 57386
const LOWERCASE_IDENTIFIER = 57387
const MACRO = 57388
const MANDATORY_GROUPS = 57389
const MAX_ACCESS = 57390
const MIN_ACCESS = 57391
const MODULE = 57392
const MODULE_COMPLIANCE = 57393
const MODULE_IDENTITY = 57394
const NEGATIVE_NUMBER = 57395
const NEGATIVE_NUMBER64 = 57396
const NOTIFICATION_GROUP = 57397
const NOTIFICATION_TYPE = 57398
const NOTIFICATIONS = 57399
const NUMBER = 57400
const NUMBER64 = 57401
const OBJECT = 57402
const OBJECT_GROUP = 57403
const OBJECT_IDENTITY = 57404
const OBJECT_TYPE = 57405
const OBJECTS = 57406
const OCTET = 57407
const OF = 57408
const ORGANIZATION = 57409
const OPAQUE = 57410
const PIB_ACCESS = 57411
const PIB_DEFINITIONS = 57412
const PIB_INDEX = 57413
const PIB_MIN_ACCESS = 57414
const PIB_REFERENCES = 57415
const PIB_TAG = 57416
const POLICY_ACCESS = 57417
const PRODUCT_RELEASE = 57418
const QUOTED_STRING = 57419
const REFERENCE = 57420
const REVISION = 57421
const SEQUENCE = 57422
const SIZE = 57423
const SPECIAL_CHAR = 57424
const STATUS = 57425
const STRING = 57426
const SUBJECT_CATEGORIES = 57427
const SUPPORTS = 57428
const SYNTAX = 57429
const TEXTUAL_CONVENTION = 57430
const TIMETICKS = 57431
const TRAP_TYPE = 57432
const UNIQUENESS = 57433
const UNITS = 57434
const UNIVERSAL = 57435
const UNSIGNED32 = 57436
const UNSIGNED64 = 57437
const UPPERCASE_IDENTIFIER = 57438
const VALUE = 57439
const VARIABLES = 57440
const VARIATION = 57441
const WRITE_SYNTAX = 57442

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"ACCESS",
	"AGENT_CAPABILITIES",
	"APPLICATION",
	"AUGMENTS",
	"BEGIN",
	"BIN_STRING",
	"BITS",
	"CHOICE",
	"COLON_COLON_EQUAL",
	"COMMENT",
	"CONTACT_INFO",
	"CREATION_REQUIRES",
	"COUNTER",
	"COUNTER32",
	"COUNTER64",
	"DEFINITIONS",
	"DEFVAL",
	"DESCRIPTION",
	"DISPLAY_HINT",
	"DOT_DOT",
	"END",
	"ENTERPRISE",
	"EXPORTS",
	"EXTENDS",
	"FROM",
	"GROUP",
	"GAUGE",
	"GAUGE32",
	"HEX_STRING",
	"IDENTIFIER",
	"IMPLICIT",
	"IMPLIED",
	"IMPORTS",
	"INCLUDES",
	"INDEX",
	"INSTALL_ERRORS",
	"INTEGER",
	"INTEGER32",
	"INTEGER64",
	"IPADDRESS",
	"LAST_UPDATED",
	"LOWERCASE_IDENTIFIER",
	"MACRO",
	"MANDATORY_GROUPS",
	"MAX_ACCESS",
	"MIN_ACCESS",
	"MODULE",
	"MODULE_COMPLIANCE",
	"MODULE_IDENTITY",
	"NEGATIVE_NUMBER",
	"NEGATIVE_NUMBER64",
	"NOTIFICATION_GROUP",
	"NOTIFICATION_TYPE",
	"NOTIFICATIONS",
	"NUMBER",
	"NUMBER64",
	"OBJECT",
	"OBJECT_GROUP",
	"OBJECT_IDENTITY",
	"OBJECT_TYPE",
	"OBJECTS",
	"OCTET",
	"OF",
	"ORGANIZATION",
	"OPAQUE",
	"PIB_ACCESS",
	"PIB_DEFINITIONS",
	"PIB_INDEX",
	"PIB_MIN_ACCESS",
	"PIB_REFERENCES",
	"PIB_TAG",
	"POLICY_ACCESS",
	"PRODUCT_RELEASE",
	"QUOTED_STRING",
	"REFERENCE",
	"REVISION",
	"SEQUENCE",
	"SIZE",
	"SPECIAL_CHAR",
	"STATUS",
	"STRING",
	"SUBJECT_CATEGORIES",
	"SUPPORTS",
	"SYNTAX",
	"TEXTUAL_CONVENTION",
	"TIMETICKS",
	"TRAP_TYPE",
	"UNIQUENESS",
	"UNITS",
	"UNIVERSAL",
	"UNSIGNED32",
	"UNSIGNED64",
	"UPPERCASE_IDENTIFIER",
	"VALUE",
	"VARIABLES",
	"VARIATION",
	"WRITE_SYNTAX",
	"'{'",
	"'}'",
	"';'",
	"','",
	"'('",
	"')'",
	"'['",
	"']'",
	"'.'",
	"'|'",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line parser.y:1075

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 19,
	109, 49,
	-2, 80,
	-1, 32,
	24, 51,
	-2, 0,
	-1, 38,
	24, 50,
	-2, 0,
	-1, 57,
	12, 83,
	-2, 80,
	-1, 149,
	109, 49,
	-2, 102,
}

const yyPrivate = 57344

const yyLast = 747

var yyAct = [...]int{

	269, 631, 230, 559, 607, 578, 516, 563, 565, 500,
	556, 522, 270, 544, 511, 489, 271, 480, 377, 386,
	464, 12, 16, 4, 451, 4, 282, 420, 135, 381,
	281, 218, 275, 258, 237, 247, 239, 268, 267, 204,
	238, 141, 144, 206, 235, 205, 86, 403, 290, 139,
	140, 291, 291, 139, 196, 203, 156, 161, 23, 301,
	156, 161, 300, 628, 570, 424, 415, 190, 402, 401,
	157, 195, 190, 114, 157, 31, 189, 619, 531, 147,
	148, 162, 155, 147, 148, 162, 155, 634, 580, 635,
	581, 539, 529, 540, 530, 502, 408, 503, 409, 152,
	347, 343, 348, 152, 151, 296, 295, 160, 151, 289,
	345, 160, 339, 593, 338, 302, 339, 303, 293, 154,
	294, 618, 287, 154, 288, 132, 212, 136, 159, 189,
	30, 24, 159, 158, 163, 149, 594, 158, 163, 149,
	598, 150, 342, 130, 35, 627, 153, 615, 614, 595,
	153, 613, 602, 597, 569, 179, 568, 590, 592, 567,
	550, 520, 589, 591, 484, 150, 483, 482, 477, 454,
	376, 341, 201, 184, 117, 21, 584, 575, 217, 561,
	180, 596, 543, 528, 185, 233, 527, 187, 191, 193,
	150, 188, 306, 192, 526, 194, 220, 208, 209, 315,
	320, 133, 214, 215, 513, 587, 211, 202, 229, 505,
	495, 491, 207, 316, 469, 210, 445, 213, 444, 443,
	441, 427, 310, 311, 321, 314, 193, 326, 260, 228,
	192, 226, 194, 224, 222, 183, 8, 517, 564, 557,
	262, 283, 313, 251, 255, 473, 436, 312, 18, 18,
	319, 256, 250, 266, 252, 147, 148, 277, 5, 279,
	265, 17, 17, 199, 286, 278, 327, 542, 336, 245,
	501, 318, 119, 323, 438, 152, 317, 322, 307, 167,
	151, 452, 171, 368, 197, 84, 481, 349, 428, 285,
	298, 232, 244, 227, 297, 177, 299, 225, 249, 19,
	19, 221, 120, 176, 166, 378, 231, 272, 412, 385,
	340, 186, 535, 240, 242, 332, 173, 169, 241, 243,
	467, 360, 625, 617, 508, 608, 359, 382, 379, 356,
	357, 361, 355, 353, 10, 536, 609, 364, 609, 549,
	390, 249, 425, 392, 509, 394, 383, 395, 128, 396,
	200, 468, 388, 389, 354, 400, 254, 253, 36, 331,
	27, 283, 175, 365, 421, 447, 366, 367, 398, 131,
	370, 371, 372, 373, 374, 393, 375, 391, 245, 397,
	335, 407, 129, 624, 623, 11, 334, 387, 507, 512,
	219, 276, 259, 248, 127, 124, 236, 351, 126, 123,
	26, 244, 422, 223, 125, 121, 122, 475, 525, 413,
	382, 15, 39, 492, 416, 417, 34, 430, 363, 423,
	362, 369, 240, 242, 404, 405, 198, 241, 243, 164,
	29, 165, 178, 115, 292, 182, 629, 551, 514, 426,
	439, 437, 499, 456, 54, 458, 442, 434, 399, 448,
	54, 116, 352, 346, 283, 344, 459, 460, 461, 476,
	440, 150, 337, 325, 284, 263, 560, 453, 457, 487,
	621, 414, 574, 14, 496, 471, 470, 433, 432, 431,
	486, 429, 490, 410, 485, 406, 493, 22, 216, 118,
	20, 25, 612, 494, 630, 622, 3, 497, 498, 6,
	620, 611, 515, 555, 504, 554, 521, 472, 450, 449,
	537, 534, 466, 465, 463, 488, 490, 533, 523, 519,
	462, 446, 435, 419, 150, 418, 532, 545, 545, 545,
	518, 605, 13, 280, 174, 172, 479, 606, 604, 585,
	150, 546, 547, 562, 566, 548, 538, 350, 358, 246,
	603, 588, 558, 523, 571, 552, 553, 586, 573, 309,
	308, 146, 579, 142, 274, 566, 572, 273, 478, 170,
	168, 510, 577, 582, 576, 333, 330, 329, 380, 324,
	261, 541, 566, 583, 524, 601, 506, 474, 455, 599,
	600, 411, 384, 105, 328, 264, 305, 234, 92, 579,
	138, 304, 257, 145, 610, 106, 107, 143, 181, 71,
	616, 70, 59, 58, 134, 53, 137, 56, 51, 108,
	50, 49, 48, 47, 626, 46, 45, 44, 632, 93,
	112, 94, 633, 87, 43, 95, 632, 636, 42, 96,
	97, 41, 40, 109, 110, 38, 91, 90, 89, 98,
	99, 100, 52, 85, 83, 69, 101, 82, 33, 37,
	32, 28, 9, 7, 2, 1, 0, 79, 81, 0,
	0, 0, 0, 0, 0, 0, 102, 103, 111, 0,
	0, 80, 104, 113, 88, 0, 0, 0, 0, 0,
	0, 77, 72, 74, 0, 55, 0, 0, 0, 0,
	0, 68, 60, 0, 0, 67, 63, 0, 0, 0,
	0, 66, 64, 61, 0, 0, 0, 0, 76, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 65, 75,
	62, 0, 0, 0, 78, 73, 57,
}
var yyPact = [...]int{

	162, -1000, 162, -1000, 135, -1000, -1000, 315, 204, 478,
	-1000, -1000, 73, 204, -1000, -1000, -51, -1000, 26, -1000,
	483, -1000, -1000, 355, 302, 404, 25, -31, 380, 41,
	300, -1000, 650, -1000, 588, -1000, -33, 409, 650, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 72, 477, 212, 343, 336, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 40, 588, -1000, 97, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 39, 396,
	406, 221, 192, 253, 197, 252, 305, 220, 219, 408,
	-1000, -1000, 162, 588, -1000, -1000, 413, -1000, -1000, 134,
	71, -1000, 215, -1000, -1000, -1000, -1000, -29, 24, -34,
	-55, 200, 393, 257, 106, -34, -34, 24, 24, -34,
	21, -34, 24, 24, 476, 204, 345, 43, 218, 133,
	359, 132, 214, 130, 210, 128, 345, 229, -1000, -1000,
	-1000, 208, 229, 351, -1000, -1000, -34, -1000, -1000, 369,
	348, -1000, -1000, -1000, -1000, 260, 156, 21, -34, 299,
	298, 155, 347, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 217, -1000, -1000, -1000, 127, 142, 444, -1000,
	168, 345, 204, 230, 346, 345, 204, 345, 204, 443,
	206, -1000, 345, -1000, 20, -1000, 4, -58, -1000, 411,
	-1000, -1000, -1000, -1000, -1000, -1000, 16, -1000, 1, 0,
	-34, -1000, -1000, -46, -49, -1000, -1000, 13, -1000, 182,
	204, 442, 126, 229, 311, 229, 441, 12, -1000, -1000,
	-1000, 243, -1000, 69, 38, -1000, -4, 434, 8, 432,
	-2, -1000, -1000, -1000, 229, 352, 431, -1000, 351, 296,
	-1000, 369, 369, -1000, 348, 268, 369, -1000, -1000, -1000,
	386, 384, -1000, 347, -1000, -1000, -1000, -34, -1000, -1000,
	-34, -34, 199, 388, -34, -34, -34, -34, -34, -1000,
	-34, -1000, -1000, 68, 227, 229, 204, 227, 236, -1000,
	-1000, 342, 342, 342, -1000, -1000, -1000, 229, -1000, 204,
	229, -1000, 346, 287, 229, -1000, 229, -1000, 204, 227,
	427, -1000, 229, -1000, -37, -1000, -1000, -1000, -38, -1000,
	-1000, -59, -1000, -1000, -1000, -1000, -1000, -1000, -34, -34,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 473, 229, -1000,
	-6, -1000, -1000, 471, 234, -1000, -1000, -1000, -1000, -1000,
	227, -1000, 457, -1000, -40, 227, 227, -1000, 314, 229,
	227, -1000, -1000, -41, -1000, -1000, 284, -1000, -1000, 204,
	120, 205, -1000, 469, 229, -1000, 467, 466, 465, 314,
	-1000, 150, 227, 187, -1000, -1000, -1000, 204, 345, 119,
	425, 118, 117, 115, -1000, 318, 204, 195, 43, 67,
	422, 204, 229, 204, 204, 204, 291, 113, -1000, 464,
	195, -1000, 149, -1000, -1000, 368, 229, 66, 207, 65,
	64, 62, -1000, 291, -1000, -1000, -1000, 204, 204, 204,
	110, -1000, 376, 204, 227, 109, -1000, -1000, 462, 207,
	-1000, 230, -1000, -1000, -1000, -1000, 421, 183, -7, -1000,
	-1000, 204, 108, -1000, 317, 344, 103, -1000, 417, 229,
	137, 43, -1000, 204, 59, 204, 370, 93, 85, 82,
	-10, -1000, -27, 204, 229, -1000, 263, 43, -1000, -1000,
	-1000, -11, -1000, -1000, 176, 81, 204, 204, 204, -1000,
	344, 281, 58, -1000, 416, 342, 342, -1000, -1000, 140,
	204, 446, 78, 203, 57, -1000, 54, 52, -1000, -42,
	-1000, 229, -1000, -1000, -1000, 140, -1000, 204, -1000, 460,
	76, 204, -14, -1000, 204, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 183, 75, 104, 51, 36, -1000, -1000,
	-1000, 203, -1000, 137, 204, 50, -1000, 280, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 204, -1000,
	488, 49, -1000, 46, 45, 278, 17, -1000, -28, -1000,
	-1000, 455, 339, -1000, -1000, -1000, -1000, -28, 338, 264,
	446, 44, -1000, -1000, -1000, -43, 415, 204, -1000, 229,
	-15, -1000, -1000, -1000, -1000, 204, -1000,
}
var yyPgo = [...]int{

	0, 665, 664, 496, 22, 663, 662, 661, 660, 659,
	12, 658, 657, 654, 285, 653, 46, 648, 647, 646,
	645, 412, 642, 641, 638, 634, 627, 626, 625, 623,
	622, 621, 620, 618, 617, 616, 411, 615, 614, 613,
	612, 611, 609, 28, 608, 31, 2, 18, 607, 42,
	603, 602, 33, 601, 600, 597, 596, 55, 44, 595,
	594, 592, 591, 588, 587, 586, 584, 581, 3, 0,
	580, 579, 578, 29, 577, 576, 575, 19, 574, 572,
	5, 571, 14, 570, 26, 569, 16, 568, 567, 564,
	32, 41, 563, 561, 560, 559, 557, 551, 39, 43,
	45, 550, 34, 40, 36, 549, 35, 548, 547, 13,
	543, 7, 8, 539, 538, 537, 536, 17, 38, 535,
	37, 534, 533, 30, 532, 473, 531, 4, 525, 523,
	27, 522, 521, 520, 515, 15, 514, 20, 513, 512,
	9, 6, 511, 510, 509, 508, 24, 507, 506, 505,
	11, 503, 10, 501, 500, 495, 494, 1,
}
var yyR1 = [...]int{

	0, 1, 1, 2, 2, 3, 5, 5, 6, 6,
	8, 8, 11, 7, 7, 12, 12, 13, 13, 14,
	15, 15, 16, 16, 16, 17, 17, 17, 17, 17,
	17, 17, 17, 17, 17, 17, 17, 17, 17, 17,
	18, 18, 18, 18, 18, 18, 18, 19, 19, 4,
	9, 9, 20, 20, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 33, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 35, 36,
	36, 23, 22, 37, 37, 37, 39, 39, 41, 41,
	41, 41, 41, 42, 42, 42, 40, 40, 38, 38,
	38, 48, 49, 50, 51, 51, 52, 43, 43, 53,
	53, 53, 55, 55, 58, 24, 25, 63, 63, 26,
	70, 70, 72, 72, 73, 71, 71, 60, 60, 60,
	75, 76, 76, 61, 61, 62, 62, 67, 67, 78,
	78, 79, 79, 80, 64, 64, 81, 81, 82, 74,
	74, 27, 28, 85, 85, 88, 89, 89, 90, 90,
	54, 54, 54, 54, 54, 54, 92, 92, 56, 56,
	96, 91, 91, 91, 91, 91, 91, 91, 91, 91,
	91, 91, 91, 91, 91, 97, 97, 97, 97, 97,
	97, 97, 97, 97, 94, 94, 94, 94, 93, 93,
	93, 93, 93, 93, 93, 93, 93, 93, 93, 93,
	93, 93, 95, 95, 95, 95, 95, 95, 95, 95,
	95, 57, 57, 57, 57, 98, 100, 102, 102, 103,
	103, 104, 104, 104, 104, 104, 104, 99, 105, 105,
	106, 107, 107, 45, 108, 44, 44, 59, 59, 77,
	65, 65, 65, 65, 66, 66, 110, 110, 111, 111,
	112, 109, 68, 68, 113, 113, 114, 114, 115, 115,
	69, 84, 47, 47, 87, 87, 116, 116, 117, 83,
	83, 119, 118, 118, 120, 121, 122, 122, 123, 46,
	86, 10, 124, 124, 125, 125, 125, 125, 125, 101,
	126, 126, 127, 127, 30, 31, 29, 128, 129, 129,
	130, 131, 131, 131, 132, 132, 134, 134, 135, 133,
	133, 136, 136, 137, 137, 138, 139, 140, 140, 141,
	141, 143, 142, 142, 142, 32, 144, 144, 145, 145,
	146, 148, 148, 150, 147, 147, 149, 149, 151, 151,
	152, 153, 153, 155, 154, 154, 156, 156, 157,
}
var yyR2 = [...]int{

	0, 1, 0, 1, 2, 9, 3, 0, 1, 1,
	1, 0, 3, 0, 2, 1, 0, 1, 2, 3,
	1, 3, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 0, 1, 2, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 2, 3, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 2, 1,
	1, 7, 3, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 9,
	1, 3, 1, 4, 1, 3, 2, 1, 4, 1,
	1, 2, 1, 3, 4, 11, 21, 2, 0, 9,
	4, 0, 1, 3, 1, 2, 0, 1, 1, 0,
	2, 1, 1, 1, 0, 1, 0, 4, 0, 1,
	0, 1, 3, 1, 4, 0, 1, 3, 4, 2,
	2, 12, 16, 4, 0, 1, 1, 3, 1, 4,
	1, 2, 1, 1, 1, 1, 5, 5, 1, 1,
	1, 1, 2, 2, 1, 2, 2, 4, 2, 4,
	2, 3, 2, 4, 3, 1, 1, 1, 1, 1,
	1, 1, 1, 3, 2, 2, 3, 3, 2, 2,
	1, 2, 1, 2, 2, 1, 2, 2, 1, 2,
	1, 2, 2, 2, 2, 2, 2, 1, 2, 1,
	1, 1, 1, 1, 0, 3, 6, 1, 3, 1,
	3, 1, 1, 1, 1, 1, 1, 3, 1, 3,
	4, 1, 1, 1, 1, 2, 0, 2, 0, 1,
	4, 4, 4, 0, 4, 0, 1, 3, 2, 1,
	1, 1, 4, 0, 1, 3, 1, 0, 1, 3,
	1, 1, 2, 0, 1, 0, 1, 2, 4, 4,
	0, 4, 1, 3, 1, 4, 1, 3, 1, 1,
	1, 1, 1, 2, 1, 3, 1, 4, 6, 1,
	1, 2, 4, 1, 12, 12, 12, 1, 1, 2,
	4, 2, 1, 0, 4, 0, 1, 3, 1, 1,
	0, 1, 2, 1, 1, 4, 7, 2, 0, 2,
	0, 1, 2, 2, 0, 14, 1, 0, 1, 2,
	7, 1, 3, 1, 2, 1, 1, 0, 1, 2,
	9, 2, 0, 1, 4, 0, 1, 3, 1,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, 96, -3, -5, 101, -6,
	19, 70, -10, -124, -125, -36, -4, 58, 45, 96,
	12, 102, -125, 109, 105, 8, 45, 58, -7, 26,
	105, 106, -8, -11, 36, 103, 58, -9, -20, -21,
	-22, -23, -24, -25, -26, -27, -28, -29, -30, -31,
	-32, -33, 2, -37, -36, 45, -34, 96, -39, -40,
	52, 63, 90, 56, 62, 88, 61, 55, 51, 5,
	-41, -42, 42, 95, 43, 89, 68, 41, 94, 17,
	31, 18, -12, -13, -14, -15, -16, 45, 96, -17,
	-18, -19, 10, 41, 43, 47, 51, 52, 61, 62,
	63, 68, 88, 89, 94, 5, 17, 18, 31, 55,
	56, 90, 42, 95, 106, 24, -21, 102, 12, 60,
	90, 62, 63, 56, 52, 61, 55, 51, 5, 46,
	103, -14, 28, 104, -38, -43, 88, -35, -54, 10,
	11, -91, -92, -48, -49, -50, -93, 40, 41, 96,
	-4, 65, 60, 107, 80, 43, 17, 31, 94, 89,
	68, 18, 42, 95, 33, 25, 83, 87, -83, 64,
	-85, 85, -119, 64, -121, 57, 83, 76, 24, -4,
	-16, -44, 22, 101, 102, -91, 96, -98, -99, 105,
	101, -98, -99, -98, -100, 105, 109, 84, 33, 6,
	93, 66, 101, -57, -98, -100, -99, -57, -98, -98,
	-57, -100, 105, -57, -98, -98, 12, -10, -45, 45,
	-43, 83, 101, 44, 101, 83, 101, 83, 101, -45,
	-46, 77, 83, -46, -55, -58, 45, -102, -103, -104,
	53, 58, 54, 59, 32, 9, -105, -106, 45, 81,
	96, -100, -57, 58, 58, -49, 96, -51, -52, 45,
	101, -70, 98, 21, -59, 92, -45, -118, -120, -69,
	-10, -86, 77, -88, -89, -90, 45, -45, -118, -45,
	-122, -123, -84, -10, 21, 83, -45, 102, 104, 105,
	106, 110, 23, 102, 104, 105, 105, -99, -98, -100,
	108, 108, 102, 104, -53, -56, 10, 96, -94, -95,
	40, 41, 65, 60, 43, 17, 31, 94, 89, 68,
	18, 42, 95, -10, -71, 21, 101, -46, -60, -74,
	-75, 48, 4, -76, 75, 69, -46, 21, 102, 104,
	67, 102, 104, 105, 21, 102, 21, 102, 104, -46,
	-108, 45, 21, -58, 58, -103, -104, -106, -107, 58,
	53, -102, 34, 34, -52, -57, -57, -57, 84, 33,
	-57, -57, -57, -57, -57, -57, 102, -47, 78, -46,
	-72, -73, -69, -47, -61, 73, -77, 45, -77, -77,
	-46, -120, -46, -90, 58, -46, -46, -123, -47, 21,
	-46, 106, 106, 106, -57, -57, 12, -46, 102, 104,
	12, -62, 74, -47, 14, 106, -47, -47, -128, -129,
	-130, 50, -46, -47, 106, 58, -73, 101, 83, 12,
	-46, 12, 12, 12, -130, -131, 96, -47, 87, -10,
	-45, 101, 21, 101, 101, 101, -132, 47, -10, -144,
	-145, -146, 86, -43, 102, -63, 21, -84, -46, -10,
	-10, -10, -133, -136, -137, -138, -139, 29, 60, 101,
	12, -146, -147, 96, -64, 39, -46, 102, -87, -116,
	-117, 79, 102, 102, 102, -137, -10, -69, -134, -135,
	-10, 101, 37, -10, -47, 101, 12, -117, -86, 21,
	-140, 87, 102, 104, -10, 101, -65, 71, 7, 27,
	-81, -82, 45, 101, 21, -46, -141, 100, -43, -135,
	102, -148, -150, -10, -66, 38, 101, 101, 101, 102,
	104, 105, -10, -46, -142, 49, 72, -143, -43, 102,
	104, -67, 91, 101, -109, -69, -109, -109, -82, 58,
	102, 21, -77, -77, -149, -151, -152, 99, -150, -68,
	20, 101, -110, -111, 35, -112, -69, 102, 102, 102,
	106, -46, -152, -69, 12, 101, -78, -79, -80, -69,
	102, 104, -112, -140, 101, -113, -96, 101, -97, 58,
	53, 59, 54, 9, 32, 45, 77, 102, 104, -111,
	-141, -69, 102, -101, -114, -126, -115, -127, 45, 58,
	-80, -153, 4, 102, 102, 102, -127, 45, 104, 105,
	-154, 15, -155, 45, 45, 58, -68, 101, 106, 21,
	-156, -157, -69, -46, 102, 104, -157,
}
var yyDef = [...]int{

	2, -2, 1, 3, 7, 49, 4, 0, 0, 0,
	8, 9, 0, 291, 292, 294, 0, 296, 79, -2,
	0, 6, 293, 0, 0, 13, 295, 0, 11, 0,
	0, 297, -2, 10, 16, 14, 0, 0, -2, 52,
	54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
	64, 65, 0, 0, 0, 79, 0, -2, 84, 85,
	68, 69, 70, 71, 72, 73, 74, 75, 76, 77,
	86, 87, 96, 97, 88, 89, 90, 91, 92, 93,
	94, 95, 0, 15, 17, 0, 20, 22, 23, 24,
	25, 26, 27, 28, 29, 30, 31, 32, 33, 34,
	35, 36, 37, 38, 39, 40, 41, 42, 43, 44,
	45, 46, 47, 48, 298, 5, 53, 66, 0, 0,
	0, 0, 0, 280, 154, 0, 0, 0, 0, 0,
	12, 18, 0, 0, 82, 98, 246, 100, 107, 0,
	0, 160, 0, 162, 163, 164, 165, 171, 174, -2,
	0, 0, 0, 0, 0, 224, 224, 200, 202, 224,
	205, 224, 208, 210, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 67, 19,
	21, 0, 0, 0, 78, 161, 49, 172, 173, 0,
	0, 175, 176, 178, 182, 0, 0, 180, 224, 0,
	0, 0, 0, 198, 221, 222, 223, 199, 201, 203,
	204, 206, 0, 207, 209, 211, 0, 121, 0, 243,
	248, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 289, 0, 245, 0, 112, 0, 0, 227, 229,
	231, 232, 233, 234, 235, 236, 0, 238, 0, 0,
	0, 181, 184, 0, 0, 101, 102, 0, 104, 0,
	0, 126, 0, 0, 129, 0, 0, 0, 282, 284,
	270, 0, 290, 0, 155, 156, 158, 0, 0, 0,
	0, 286, 288, 271, 0, 0, 0, 108, 0, 0,
	225, 0, 0, 237, 0, 0, 0, 177, 179, 183,
	0, 0, 103, 0, 106, 109, 110, 224, 168, 169,
	224, 224, 0, 0, 224, 224, 224, 224, 224, 217,
	224, 219, 220, 0, 273, 0, 0, 273, 134, 127,
	128, 0, 0, 0, 131, 132, 247, 0, 279, 0,
	0, 153, 0, 0, 0, 281, 0, 285, 0, 273,
	0, 244, 0, 113, 0, 228, 230, 239, 0, 241,
	242, 0, 166, 167, 105, 111, 194, 195, 224, 224,
	212, 213, 214, 215, 216, 218, 81, 0, 0, 125,
	0, 122, 124, 0, 136, 133, 149, 249, 150, 130,
	273, 283, 0, 157, 0, 273, 273, 287, 0, 0,
	273, 114, 240, 0, 196, 197, 0, 272, 120, 0,
	0, 0, 135, 0, 0, 159, 0, 0, 0, 307,
	308, 313, 273, 0, 226, 119, 123, 0, 0, 0,
	0, 0, 0, 0, 309, 315, 312, 337, 0, 0,
	118, 0, 0, 0, 0, 0, 320, 0, 311, 0,
	336, 338, 0, 99, 115, 145, 0, 0, 275, 0,
	0, 0, 310, 319, 321, 323, 324, 0, 0, 0,
	0, 339, 0, 345, 273, 0, 117, 151, 0, 274,
	276, 0, 304, 305, 306, 322, 0, 328, 0, 316,
	318, 0, 0, 344, 253, 0, 0, 277, 0, 0,
	330, 0, 314, 0, 0, 0, 255, 0, 0, 0,
	0, 146, 0, 0, 0, 325, 334, 0, 327, 317,
	335, 0, 341, 343, 138, 0, 0, 0, 0, 144,
	0, 0, 0, 278, 0, 0, 0, 329, 331, 347,
	0, 263, 0, 0, 0, 261, 0, 0, 147, 0,
	152, 0, 332, 333, 340, 346, 348, 0, 342, 0,
	0, 140, 0, 256, 0, 259, 260, 250, 251, 252,
	148, 326, 349, 328, 0, 0, 0, 139, 141, 143,
	254, 0, 258, 330, 0, 0, 264, 267, 170, 185,
	186, 187, 188, 189, 190, 191, 192, 137, 0, 257,
	352, 0, 262, 0, 0, 299, 266, 300, 268, 303,
	142, 355, 0, 116, 193, 265, 301, 0, 0, 0,
	263, 0, 351, 353, 269, 0, 0, 0, 302, 0,
	0, 356, 358, 350, 354, 0, 357,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	105, 106, 3, 3, 104, 3, 109, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 103,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 107, 3, 108, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 101, 110, 102,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
	52, 53, 54, 55, 56, 57, 58, 59, 60, 61,
	62, 63, 64, 65, 66, 67, 68, 69, 70, 71,
	72, 73, 74, 75, 76, 77, 78, 79, 80, 81,
	82, 83, 84, 85, 86, 87, 88, 89, 90, 91,
	92, 93, 94, 95, 96, 97, 98, 99, 100,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:161
		{
			// Add modules to the module map stored in the lexer
			for _, m := range yyVAL.modules {
				yylex.(*lexer).modules[m.name] = m
			}
			// Clear object, module data from yys
			yyVAL.objectMap = make(map[string]*parseObject)
			yyVAL.orphans = []*parseObject{}
			yyVAL.objects = []*parseObject{}
			yyVAL.modules = []*parseModule{}
		}
	case 5:
		yyDollar = yyS[yypt-9 : yypt+1]
//line parser.y:180
		{
			m := &parseModule{
				imports:    yyDollar[7].imports,
				name:       yyDollar[1].val,
				objectTree: []*parseObject{},
				orphans:    []*parseObject{},
			}
			for _, o := range yyDollar[8].objects {
				m.objectTree = append(m.objectTree, o)
				o.setModule(m.name)
			}
			for _, o := range yyDollar[8].orphans {
				m.orphans = append(m.orphans, o)
			}
			yyVAL.addModule(m)
		}
	case 10:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:207
		{
			yyVAL.imports = yyDollar[1].imports
		}
	case 11:
		yyDollar = yyS[yypt-0 : yypt+1]
//line parser.y:211
		{
			yyVAL.imports = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:217
		{
			yyVAL.imports = yyDollar[2].imports
		}
	case 15:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:227
		{
			yyVAL.imports = yyDollar[1].imports
		}
	case 16:
		yyDollar = yyS[yypt-0 : yypt+1]
//line parser.y:231
		{
			yyVAL.imports = nil
		}
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:237
		{
			yyVAL.imports = yyDollar[1].imports
		}
	case 18:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:241
		{
			yyVAL.imports = append(yyDollar[1].imports, yyDollar[2].imports...)
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:247
		{
			yyVAL.imports = []Import{}
			for _, id := range yyDollar[1].importIDs {
				yyVAL.imports = append(yyVAL.imports,
					Import{Object: id, Module: yyDollar[3].token.literal})
			}
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:257
		{
			yyVAL.importIDs = []string{yyDollar[1].token.literal}
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:261
		{
			yyVAL.importIDs = append(yyDollar[1].importIDs, yyDollar[3].token.literal)
		}
	case 49:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:302
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 52:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:312
		{
			(&yyVAL).addObject(yyDollar[1].object)
		}
	case 53:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:316
		{
			(&yyVAL).addObject(yyDollar[2].object)
		}
	case 54:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:322
		{
			(&yyVAL).setDecl(declTypeAssignment)
		}
	case 55:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:326
		{
			(&yyVAL).setDecl(declValueAssignment)
		}
	case 56:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:330
		{
			(&yyVAL).setDecl(declIdentity)
		}
	case 57:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:334
		{
			(&yyVAL).setDecl(declObjectType)
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:338
		{
			(&yyVAL).setDecl(declTrapType)
		}
	case 59:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:342
		{
			(&yyVAL).setDecl(declNotificationType)
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:346
		{
			(&yyVAL).setDecl(declModuleIdentity)
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:350
		{
			(&yyVAL).setDecl(declModuleCompliance)
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:354
		{
			(&yyVAL).setDecl(declObjectGroup)
		}
	case 63:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:358
		{
			(&yyVAL).setDecl(declNotificationGroup)
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:362
		{
			(&yyVAL).setDecl(declAgentCapabilities)
		}
	case 79:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:388
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 80:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:392
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 81:
		yyDollar = yyS[yypt-7 : yypt+1]
//line parser.y:398
		{
			yyVAL.object = &parseObject{
				object: &Object{
					Name: yyDollar[1].val,
					Oid:  strings.Join(yyDollar[6].subidentifiers, "."),
				},
			}
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:437
		{
			yyVAL.table = yyDollar[1].table
		}
	case 99:
		yyDollar = yyS[yypt-9 : yypt+1]
//line parser.y:441
		{
			yyVAL.table = yyDollar[9].table
			yyVAL.status = strToStatus(yyDollar[4].val)
			yyVAL.description = yyDollar[6].val
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:450
		{
			yyVAL.table = true
		}
	case 115:
		yyDollar = yyS[yypt-11 : yypt+1]
//line parser.y:485
		{
			yyVAL.object = &parseObject{
				object: &Object{
					Description: yyDollar[6].val,
					Name:        yyDollar[1].token.literal,
					Oid:         strings.Join(yyDollar[10].subidentifiers, "."),
					Status:      strToStatus(yyDollar[4].val),
				},
				decl: declIdentity,
			}
		}
	case 116:
		yyDollar = yyS[yypt-21 : yypt+1]
//line parser.y:499
		{
			yyVAL.object = &parseObject{
				object: &Object{
					Access:      strToAccess(yyDollar[6].val),
					Description: yyDollar[11].val,
					Indexes:     yyDollar[15].indexes,
					Name:        yyDollar[1].token.literal,
					Oid:         strings.Join(yyDollar[20].subidentifiers, "."),
					Status:      strToStatus(yyDollar[10].val),
				},
				decl:     declObjectType,
				table:    yyDollar[4].table,
				augments: yyDollar[14].augments,
			}
		}
	case 117:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:517
		{
			yyVAL.val = yyDollar[2].val
		}
	case 149:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:588
		{
			yyVAL.val = yyDollar[2].token.literal
		}
	case 150:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:592
		{
			yyVAL.val = yyDollar[2].token.literal
		}
	case 152:
		yyDollar = yyS[yypt-16 : yypt+1]
//line parser.y:602
		{
			yyVAL.object = &parseObject{
				object: &Object{
					Name:        yyDollar[1].token.literal,
					Oid:         strings.Join(yyDollar[15].subidentifiers, "."),
					Description: yyDollar[11].val,
				},
			}
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:750
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 250:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:770
		{
			yyVAL.augments = ""
		}
	case 251:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:774
		{
			yyVAL.augments = yyDollar[3].subidentifiers[0]
		}
	case 252:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:778
		{
			yyVAL.augments = ""
		}
	case 253:
		yyDollar = yyS[yypt-0 : yypt+1]
//line parser.y:782
		{
			yyVAL.augments = ""
		}
	case 254:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:788
		{
			yyVAL.indexes = yyDollar[3].indexes
		}
	case 255:
		yyDollar = yyS[yypt-0 : yypt+1]
//line parser.y:792
		{
			yyVAL.indexes = nil
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:798
		{
			if yyDollar[1].val != "" {
				yyVAL.indexes = []string{yyDollar[1].val}
			}
		}
	case 257:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:804
		{
			if yyDollar[3].val != "" {
				yyVAL.indexes = append(yyDollar[1].indexes, yyDollar[3].val)
			}
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:813
		{
			yyVAL.val = strings.Join(yyDollar[1].subidentifiers, " ")
		}
	case 289:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:886
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 292:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:898
		{
			yyVAL.subidentifiers = []string{yyDollar[1].val}
		}
	case 293:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:902
		{
			yyVAL.subidentifiers = append(yyDollar[1].subidentifiers, yyDollar[2].val)
		}
	case 294:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:908
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 295:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:912
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 296:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:916
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 297:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:920
		{
			yyVAL.val = yyDollar[3].token.literal
		}
	case 298:
		yyDollar = yyS[yypt-6 : yypt+1]
//line parser.y:924
		{
			yyVAL.val = yyDollar[1].token.literal
		}
	case 304:
		yyDollar = yyS[yypt-12 : yypt+1]
//line parser.y:941
		{
			// XXX TODO
		}
	case 305:
		yyDollar = yyS[yypt-12 : yypt+1]
//line parser.y:947
		{
			// XXX TODO
		}
	case 306:
		yyDollar = yyS[yypt-12 : yypt+1]
//line parser.y:953
		{
			/// XXX TODO
		}
	case 335:
		yyDollar = yyS[yypt-14 : yypt+1]
//line parser.y:1019
		{
			// XXX TODO
		}
	}
	goto yystack /* stack new state and value */
}
