#ifndef CONVERTERX_H
#define CONVERTERX_H
#include <string>
#include <locale>
#include <codecvt>

std::wstring wstringFromUTF8(std::string text);
std::wstring wstringFromUTF8(char* text);

#endif
