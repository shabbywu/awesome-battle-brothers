#include "../utils/converterX.h"

using convert_typeX = std::codecvt_utf8<wchar_t>;
std::wstring_convert<convert_typeX, wchar_t> converterX;

std::wstring wstringFromUTF8(std::string text) {
    return converterX.from_bytes(text);
};

std::wstring wstringFromUTF8(char* text) {
    return converterX.from_bytes(text);
};
