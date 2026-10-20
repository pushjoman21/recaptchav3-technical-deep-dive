// Authored fixture. Parsed as data by the Go tool; never executed.
var table = [function (start, count, key, cached) {
  return function (i, text) {
    if (cached) return cached;
    for (text = "", i = 0; i < count; i++)
      text += String.fromCharCode("abcdefghijklmnopqrstuvwxyz0123456789".charCodeAt(start + i) ^ key);
    return cached = text;
  };
}];
var message = table[0](1, 3, 0);
message();
