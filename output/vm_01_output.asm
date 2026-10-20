; Redacted full research listing; all recovered handler records are retained.
; vm_01 stack handlers; IDs are registration metadata.
; [REDACTED: source identity]
.unresolved CONSUMING_INTERPRETER_AND_PROGRAM

H_000:  MUL                        ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left * right.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_001:  LENGTH                     ; ids [REDACTED: registration mapping]
  ; Pop a value; push its length property.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_002:  NOT_01                     ; ids [REDACTED: registration mapping]
  ; Pop a value; push 0 if truthy, otherwise 1.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_003:  CLOCK_I32                  ; ids [REDACTED: registration mapping]
  ; Call clock(); push its signed 32-bit coercion. The clock provider is external.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_004:  CALL_TARGET_POP            ; ids [REDACTED: registration mapping]
  ; Pop a bit address and call [REDACTED: helper locator]: save four frame fields, set the byte cursor to address >>> 3.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_005:  REPLACE_STACK_IMMEDIATE    ; ids [REDACTED: registration mapping]
  ; Read a signed immediate index; pop replacement; read the indexed old stack value, call replaceStack(index,replacement), push old value. put is external.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_006:  LOAD_FRAME_DYNAMIC         ; ids [REDACTED: registration mapping]
  ; Pop a frame-relative index; push stack[frames[frames.length-2] + index - 1].
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_007:  DIV_I32                    ; ids [REDACTED: registration mapping]
  ; Pop right and left; push (left / right) | 0.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_008:  DELETE_STACK_DYNAMIC       ; ids [REDACTED: registration mapping]
  ; Pop index; peek the top value as a mutable array; splice its relative/absolute index.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_009:  XOR                        ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left ^ right.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_010:  PEEK_DYNAMIC               ; ids [REDACTED: registration mapping]
  ; Pop an index; push stack[(~(index >> 31) & stack.length) - index - 1].
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_011:  BRANCH_NONZERO_POP         ; ids [REDACTED: registration mapping]
  ; Pop bit target, then condition; when condition !== 0 set cursor = target >>> 3.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_012:  USHR                       ; ids [REDACTED: registration mapping]
  ; Pop shift count and value; push value >>> count.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_013:  PUSH_NULLS_IMMEDIATE       ; ids [REDACTED: registration mapping]
  ; Read signed immediate count; push null that many times. Negative counts push nothing.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_014:  PUSH_NEGATIVE_TWO          ; ids [REDACTED: registration mapping]
  ; Push -2.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_015:  APPEND_SERIALIZED          ; ids [REDACTED: registration mapping]
  ; Pop value and prefix; serialize value through [REDACTED: helper locator], then byte-string helper [REDACTED: helper locator], append and push. Numeric serialization and uncached byte conversion contain literal null callees.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_016:  PUSH_NEGATIVE_ONE          ; ids [REDACTED: registration mapping]
  ; Push -1.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_017:  REMOVE_STACK_DYNAMIC       ; ids [REDACTED: registration mapping]
  ; Pop an index; remove the corresponding stack element by shifting successors down, then push the removed value.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_018:  BROKEN_BINARY_NULL_A       ; ids [REDACTED: registration mapping]
  ; Pop right and left, call literal null(left,right), then intended push is unreachable after TypeError. Intended operator is unknown.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_019:  SUB_I32                    ; ids [REDACTED: registration mapping]
  ; Pop right and left; push (left - right) | 0.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_020:  BIT_OR                     ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left | right.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_021:  BROKEN_FLOAT64             ; ids [REDACTED: registration mapping]
  ; Attempt cipher slice using missing missing byte-view constructor constructor; then new null(buffer).getFloat64(0,true). Cursor increment and push occur only if both missing dependencies succeed.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_022:  DIV                        ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left / right.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_023:  BIT_NOT                    ; ids [REDACTED: registration mapping]
  ; Pop a value; push ~value.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_024:  APPEND_VARINT64            ; ids [REDACTED: registration mapping]
  ; Pop Number and prefix; emit low/high 32-bit unsigned varint pieces, convert with [REDACTED: helper locator], append and push. [REDACTED: helper locator] contains an uncached null.apply call.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_025:  DECRYPT_STRING             ; ids [REDACTED: registration mapping]
  ; Pop seed, 8-code-unit key text and ciphertext; construct cipher constructor/[REDACTED: helper locator] cipher state, XOR eight-round blocks, convert via [REDACTED: helper locator], push. missing byte-view constructor and missing decoder dependency are null in the supplied capture.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_026:  REM_I32                    ; ids [REDACTED: registration mapping]
  ; Pop right and left; push (left % right) | 0.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_027:  CALL_RELATIVE_IMMEDIATE    ; ids [REDACTED: registration mapping]
  ; Read signed bit displacement; call frame helper at (cursor << 3) + displacement.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_028:  GET_INDEX_DEFAULT          ; ids [REDACTED: registration mapping]
  ; Pop index, default and object; normalize negative index by object.length, replace default only if index is within bounds, push result.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_029:  PUSH_IMMEDIATE             ; ids [REDACTED: registration mapping]
  ; Read encrypted ZigZag varint and push its signed value.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_030:  RETURN_FRAME               ; ids [REDACTED: registration mapping]
  ; Update saved frame/cursor state, pop four frame fields, restore byte cursor and trim the value stack to saved length.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_031:  REMOVE_STACK_IMMEDIATE     ; ids [REDACTED: registration mapping]
  ; Read signed index; remove the indexed stack element and push it.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_032:  PUSH_NULL                  ; ids [REDACTED: registration mapping]
  ; Push null.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_033:  LESS_THAN_01               ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left < right ? 1 : 0.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_034:  PUSH_TWO                   ; ids [REDACTED: registration mapping]
  ; Push 2.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_035:  SET_INDEX                  ; ids [REDACTED: registration mapping]
  ; Pop index, value and array; assign array[(index >> 31 & array.length) + index] = value; push array.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_036:  BROKEN_BINARY_NULL_B       ; ids [REDACTED: registration mapping]
  ; Pop right and left, call literal null(left,right); no replacement operator can be inferred.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_037:  PEEK_IMMEDIATE             ; ids [REDACTED: registration mapping]
  ; Read signed index and push the corresponding absolute/relative stack value.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_038:  ADD_I32                    ; ids [REDACTED: registration mapping]
  ; Pop right and left; push (left + right) | 0, preserving + coercion before truncation.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_039:  LOAD_FRAME_IMMEDIATE       ; ids [REDACTED: registration mapping]
  ; Read signed frame-relative index; push the addressed stack value.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_040:  STORE_FRAME_IMMEDIATE      ; ids [REDACTED: registration mapping]
  ; Read signed frame-relative index, pop value, assign addressed stack slot.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_041:  NOTIFY_EXTERNAL            ; ids [REDACTED: registration mapping]
  ; Pop value and invoke external callback(value); its callback contract is absent.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_042:  PUSH_THREE                 ; ids [REDACTED: registration mapping]
  ; Push 3.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_043:  BROKEN_UNARY_NULL          ; ids [REDACTED: registration mapping]
  ; Pop value and call literal null(value); push is unreachable after TypeError.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_044:  ADD                        ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left + right with JavaScript coercions.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_045:  BIT_AND                    ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left & right.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_046:  REPLACE_STACK_DYNAMIC      ; ids [REDACTED: registration mapping]
  ; Pop index and replacement; peek old indexed value, call external replaceStack(index,replacement), push old value.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_047:  GET_PROPERTY_NULLISH       ; ids [REDACTED: registration mapping]
  ; Pop key and object; push undefined if object == null, otherwise object[key].
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_048:  READ_BYTE_STRING           ; ids [REDACTED: registration mapping]
  ; Read signed bit length, shift >>> 3, decrypt that byte slice, advance cursor, convert bytes with [REDACTED: helper locator] and push. Missing native aliases prevent a complete standalone path.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_049:  APPLY_NULLISH              ; ids [REDACTED: registration mapping]
  ; Pop argument array, function and receiver; push undefined for nullish function, otherwise fn.apply(receiver,args).
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_050:  PUSH_FOUR                  ; ids [REDACTED: registration mapping]
  ; Push 4.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_051:  TYPEOF                     ; ids [REDACTED: registration mapping]
  ; Pop value and push typeof value.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_052:  SHL                        ; ids [REDACTED: registration mapping]
  ; Pop shift count and value; push value << count.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_053:  EXCHANGE_FRAME_DYNAMIC     ; ids [REDACTED: registration mapping]
  ; Pop frame index; read old frame slot, replace it with current top stack value, call external put(0,old).
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_054:  SUB                        ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left - right.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_055:  ADD_ALIAS                  ; ids [REDACTED: registration mapping]
  ; Pop right and left; push left + right. Same core behavior as ADD, separately registered.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_056:  EQUAL_NULLISH_01           ; ids [REDACTED: registration mapping]
  ; Pop two values; push 1 for strict equality or when both are nullish, otherwise 0.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_057:  CALL_ABSOLUTE_IMMEDIATE    ; ids [REDACTED: registration mapping]
  ; Read signed bit address and invoke the four-field frame helper.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_058:  BRANCH_RELATIVE_NONZERO    ; ids [REDACTED: registration mapping]
  ; Read signed bit displacement, pop condition; if !== 0 set cursor = ((cursor << 3) + displacement) >>> 3.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_059:  PUSH_ONE                   ; ids [REDACTED: registration mapping]
  ; Push 1.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_060:  SPREAD_ITERABLE            ; ids [REDACTED: registration mapping]
  ; Pop iterable; push its values in iteration order. Preserve iterator cleanup on abrupt completion.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_061:  BRANCH_ABSOLUTE_ZERO       ; ids [REDACTED: registration mapping]
  ; Read signed bit target, pop condition; if === 0 set cursor = target >>> 3.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_062:  PUSH_ZERO                  ; ids [REDACTED: registration mapping]
  ; Push 0.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_063:  LENGTH_ALIAS               ; ids [REDACTED: registration mapping]
  ; Pop value and push its length. Same core operation as LENGTH.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_064:  MAKE_ASYNC_CLOSURE         ; ids [REDACTED: registration mapping]
  ; Pop bit entry; create a Promise-returning function that builds task records work and updates task queue priorities/deadlines through scheduler helpers and [REDACTED: helper locator]. Required scheduler members/clock are missing.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_065:  PUSH_SEVEN                 ; ids [REDACTED: registration mapping]
  ; Push 7.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_066:  CALL_DYNAMIC_FRAME         ; ids [REDACTED: registration mapping]
  ; Pop bit target and frame argument selector; call the four-field frame helper with a frame-relative base for nonzero selector.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_067:  BOOLEAN_OR_01              ; ids [REDACTED: registration mapping]
  ; Pop right and left; push (left || right) ? 1 : 0.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_068:  BOOLEAN_AND_01             ; ids [REDACTED: registration mapping]
  ; Pop right and left; push (left && right) ? 1 : 0.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_069:  BRANCH_ABSOLUTE_NONZERO    ; ids [REDACTED: registration mapping]
  ; Read signed bit target, pop condition; if !== 0 set cursor = target >>> 3.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_070:  EXCHANGE_FRAME_IMMEDIATE   ; ids [REDACTED: registration mapping]
  ; Read signed frame index, exchange addressed value with top-of-stack via external put(0,old).
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_071:  DROP                       ; ids [REDACTED: registration mapping]
  ; Pop and discard one value.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_072:  BRANCH_RELATIVE_ZERO       ; ids [REDACTED: registration mapping]
  ; Read signed bit displacement, pop condition; if === 0 set cursor = ((cursor << 3) + displacement) >>> 3.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_073:  BRANCH_ZERO_POP            ; ids [REDACTED: registration mapping]
  ; Pop bit target and condition; if condition === 0 set cursor = target >>> 3.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_074:  GET_GLOBAL_PATH            ; ids [REDACTED: registration mapping]
  ; Pop dotted path and traverse external host global root until a nullish intermediate; empty path returns root. Preserve iterator cleanup.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_075:  PACK_ARRAY                 ; ids [REDACTED: registration mapping]
  ; Pop count, pop that many values, reverse the array to restore original stack order, then push it.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

H_076:  INSERT_STACK_DYNAMIC       ; ids [REDACTED: registration mapping]
  ; Pop index and value; peek array and splice value before nonnegative index, or after the relative negative index.
  .source [REDACTED: original source locator]
  .requires [REDACTED: source binding dependencies]

