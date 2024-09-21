#pragma once
#include "physfs.override.hpp"

#include <physfs.h>

namespace patchs {
namespace physfs {
namespace stub {
void _PHYSFS_getLinkedVersion(PHYSFS_Version *ver) {
    PHYSFS_getLinkedVersion(ver);
    return;
}

const PHYSFS_ArchiveInfo **_PHYSFS_supportedArchiveTypes() {
    auto result = PHYSFS_supportedArchiveTypes();
    return result;
}

void _PHYSFS_freeList(void *listVar) {
    PHYSFS_freeList(listVar);
    return;
}

const char *_PHYSFS_getLastError() {
    auto result = PHYSFS_getLastError();
    return result;
}

const char *_PHYSFS_getDirSeparator() {
    auto result = PHYSFS_getDirSeparator();
    return result;
}

void _PHYSFS_permitSymbolicLinks(int allow) {
    PHYSFS_permitSymbolicLinks(allow);
    return;
}

char **_PHYSFS_getCdRomDirs() {
    auto result = PHYSFS_getCdRomDirs();
    return result;
}

const char *_PHYSFS_getBaseDir() {
    auto result = PHYSFS_getBaseDir();
    return result;
}

const char *_PHYSFS_getUserDir() {
    auto result = PHYSFS_getUserDir();
    return result;
}

const char *_PHYSFS_getWriteDir() {
    auto result = PHYSFS_getWriteDir();
    return result;
}

int _PHYSFS_setWriteDir(const char *newDir) {
    auto result = PHYSFS_setWriteDir(newDir);
    return result;
}

int _PHYSFS_addToSearchPath(const char *newDir, int appendToPath) {
    auto result = PHYSFS_addToSearchPath(newDir, appendToPath);
    return result;
}

int _PHYSFS_removeFromSearchPath(const char *oldDir) {
    auto result = PHYSFS_removeFromSearchPath(oldDir);
    return result;
}

char **_PHYSFS_getSearchPath() {
    auto result = PHYSFS_getSearchPath();
    return result;
}

int _PHYSFS_setSaneConfig(const char *organization, const char *appName, const char *archiveExt, int includeCdRoms,
                          int archivesFirst) {
    auto result = PHYSFS_setSaneConfig(organization, appName, archiveExt, includeCdRoms, archivesFirst);
    return result;
}

int _PHYSFS_mkdir(const char *dirName) {
    auto result = PHYSFS_mkdir(dirName);
    return result;
}

int _PHYSFS_delete(const char *filename) {
    auto result = PHYSFS_delete(filename);
    return result;
}

const char *_PHYSFS_getRealDir(const char *filename) {
    auto result = PHYSFS_getRealDir(filename);
    return result;
}

char **_PHYSFS_enumerateFiles(const char *dir) {
    auto result = PHYSFS_enumerateFiles(dir);
    return result;
}

int _PHYSFS_exists(const char *fname) {
    auto result = PHYSFS_exists(fname);
    return result;
}

int _PHYSFS_isDirectory(const char *fname) {
    auto result = PHYSFS_isDirectory(fname);
    return result;
}

int _PHYSFS_isSymbolicLink(const char *fname) {
    auto result = PHYSFS_isSymbolicLink(fname);
    return result;
}

PHYSFS_sint64 _PHYSFS_getLastModTime(const char *filename) {
    auto result = PHYSFS_getLastModTime(filename);
    return result;
}

PHYSFS_File *_PHYSFS_openWrite(const char *filename) {
    auto result = PHYSFS_openWrite(filename);
    return result;
}

PHYSFS_File *_PHYSFS_openAppend(const char *filename) {
    auto result = PHYSFS_openAppend(filename);
    return result;
}

PHYSFS_File *_PHYSFS_openRead(const char *filename) {
    auto result = PHYSFS_openRead(filename);
    return result;
}

int _PHYSFS_close(PHYSFS_File *handle) {
    auto result = PHYSFS_close(handle);
    return result;
}

PHYSFS_sint64 _PHYSFS_read(PHYSFS_File *handle, void *buffer, PHYSFS_uint32 objSize, PHYSFS_uint32 objCount) {
    auto result = PHYSFS_read(handle, buffer, objSize, objCount);
    return result;
}

PHYSFS_sint64 _PHYSFS_write(PHYSFS_File *handle, const void *buffer, PHYSFS_uint32 objSize, PHYSFS_uint32 objCount) {
    auto result = PHYSFS_write(handle, buffer, objSize, objCount);
    return result;
}

int _PHYSFS_eof(PHYSFS_File *handle) {
    auto result = PHYSFS_eof(handle);
    return result;
}

PHYSFS_sint64 _PHYSFS_tell(PHYSFS_File *handle) {
    auto result = PHYSFS_tell(handle);
    return result;
}

int _PHYSFS_seek(PHYSFS_File *handle, PHYSFS_uint64 pos) {
    auto result = PHYSFS_seek(handle, pos);
    return result;
}

PHYSFS_sint64 _PHYSFS_fileLength(PHYSFS_File *handle) {
    auto result = PHYSFS_fileLength(handle);
    return result;
}

int _PHYSFS_setBuffer(PHYSFS_File *handle, PHYSFS_uint64 bufsize) {
    auto result = PHYSFS_setBuffer(handle, bufsize);
    return result;
}

int _PHYSFS_flush(PHYSFS_File *handle) {
    auto result = PHYSFS_flush(handle);
    return result;
}

PHYSFS_sint16 _PHYSFS_swapSLE16(PHYSFS_sint16 val) {
    auto result = PHYSFS_swapSLE16(val);
    return result;
}

PHYSFS_uint16 _PHYSFS_swapULE16(PHYSFS_uint16 val) {
    auto result = PHYSFS_swapULE16(val);
    return result;
}

PHYSFS_sint32 _PHYSFS_swapSLE32(PHYSFS_sint32 val) {
    auto result = PHYSFS_swapSLE32(val);
    return result;
}

PHYSFS_uint32 _PHYSFS_swapULE32(PHYSFS_uint32 val) {
    auto result = PHYSFS_swapULE32(val);
    return result;
}

PHYSFS_sint64 _PHYSFS_swapSLE64(PHYSFS_sint64 val) {
    auto result = PHYSFS_swapSLE64(val);
    return result;
}

PHYSFS_uint64 _PHYSFS_swapULE64(PHYSFS_uint64 val) {
    auto result = PHYSFS_swapULE64(val);
    return result;
}

PHYSFS_sint16 _PHYSFS_swapSBE16(PHYSFS_sint16 val) {
    auto result = PHYSFS_swapSBE16(val);
    return result;
}

PHYSFS_uint16 _PHYSFS_swapUBE16(PHYSFS_uint16 val) {
    auto result = PHYSFS_swapUBE16(val);
    return result;
}

PHYSFS_sint32 _PHYSFS_swapSBE32(PHYSFS_sint32 val) {
    auto result = PHYSFS_swapSBE32(val);
    return result;
}

PHYSFS_uint32 _PHYSFS_swapUBE32(PHYSFS_uint32 val) {
    auto result = PHYSFS_swapUBE32(val);
    return result;
}

PHYSFS_sint64 _PHYSFS_swapSBE64(PHYSFS_sint64 val) {
    auto result = PHYSFS_swapSBE64(val);
    return result;
}

PHYSFS_uint64 _PHYSFS_swapUBE64(PHYSFS_uint64 val) {
    auto result = PHYSFS_swapUBE64(val);
    return result;
}

int _PHYSFS_readSLE16(PHYSFS_File *file, PHYSFS_sint16 *val) {
    auto result = PHYSFS_readSLE16(file, val);
    return result;
}

int _PHYSFS_readULE16(PHYSFS_File *file, PHYSFS_uint16 *val) {
    auto result = PHYSFS_readULE16(file, val);
    return result;
}

int _PHYSFS_readSBE16(PHYSFS_File *file, PHYSFS_sint16 *val) {
    auto result = PHYSFS_readSBE16(file, val);
    return result;
}

int _PHYSFS_readUBE16(PHYSFS_File *file, PHYSFS_uint16 *val) {
    auto result = PHYSFS_readUBE16(file, val);
    return result;
}

int _PHYSFS_readSLE32(PHYSFS_File *file, PHYSFS_sint32 *val) {
    auto result = PHYSFS_readSLE32(file, val);
    return result;
}

int _PHYSFS_readULE32(PHYSFS_File *file, PHYSFS_uint32 *val) {
    auto result = PHYSFS_readULE32(file, val);
    return result;
}

int _PHYSFS_readSBE32(PHYSFS_File *file, PHYSFS_sint32 *val) {
    auto result = PHYSFS_readSBE32(file, val);
    return result;
}

int _PHYSFS_readUBE32(PHYSFS_File *file, PHYSFS_uint32 *val) {
    auto result = PHYSFS_readUBE32(file, val);
    return result;
}

int _PHYSFS_readSLE64(PHYSFS_File *file, PHYSFS_sint64 *val) {
    auto result = PHYSFS_readSLE64(file, val);
    return result;
}

int _PHYSFS_readULE64(PHYSFS_File *file, PHYSFS_uint64 *val) {
    auto result = PHYSFS_readULE64(file, val);
    return result;
}

int _PHYSFS_readSBE64(PHYSFS_File *file, PHYSFS_sint64 *val) {
    auto result = PHYSFS_readSBE64(file, val);
    return result;
}

int _PHYSFS_readUBE64(PHYSFS_File *file, PHYSFS_uint64 *val) {
    auto result = PHYSFS_readUBE64(file, val);
    return result;
}

int _PHYSFS_writeSLE16(PHYSFS_File *file, PHYSFS_sint16 val) {
    auto result = PHYSFS_writeSLE16(file, val);
    return result;
}

int _PHYSFS_writeULE16(PHYSFS_File *file, PHYSFS_uint16 val) {
    auto result = PHYSFS_writeULE16(file, val);
    return result;
}

int _PHYSFS_writeSBE16(PHYSFS_File *file, PHYSFS_sint16 val) {
    auto result = PHYSFS_writeSBE16(file, val);
    return result;
}

int _PHYSFS_writeUBE16(PHYSFS_File *file, PHYSFS_uint16 val) {
    auto result = PHYSFS_writeUBE16(file, val);
    return result;
}

int _PHYSFS_writeSLE32(PHYSFS_File *file, PHYSFS_sint32 val) {
    auto result = PHYSFS_writeSLE32(file, val);
    return result;
}

int _PHYSFS_writeULE32(PHYSFS_File *file, PHYSFS_uint32 val) {
    auto result = PHYSFS_writeULE32(file, val);
    return result;
}

int _PHYSFS_writeSBE32(PHYSFS_File *file, PHYSFS_sint32 val) {
    auto result = PHYSFS_writeSBE32(file, val);
    return result;
}

int _PHYSFS_writeUBE32(PHYSFS_File *file, PHYSFS_uint32 val) {
    auto result = PHYSFS_writeUBE32(file, val);
    return result;
}

int _PHYSFS_writeSLE64(PHYSFS_File *file, PHYSFS_sint64 val) {
    auto result = PHYSFS_writeSLE64(file, val);
    return result;
}

int _PHYSFS_writeULE64(PHYSFS_File *file, PHYSFS_uint64 val) {
    auto result = PHYSFS_writeULE64(file, val);
    return result;
}

int _PHYSFS_writeSBE64(PHYSFS_File *file, PHYSFS_sint64 val) {
    auto result = PHYSFS_writeSBE64(file, val);
    return result;
}

int _PHYSFS_writeUBE64(PHYSFS_File *file, PHYSFS_uint64 val) {
    auto result = PHYSFS_writeUBE64(file, val);
    return result;
}

int _PHYSFS_isInit() {
    auto result = PHYSFS_isInit();
    return result;
}

int _PHYSFS_symbolicLinksPermitted() {
    auto result = PHYSFS_symbolicLinksPermitted();
    return result;
}

int _PHYSFS_setAllocator(const PHYSFS_Allocator *allocator) {
    auto result = PHYSFS_setAllocator(allocator);
    return result;
}

void _PHYSFS_getCdRomDirsCallback(PHYSFS_StringCallback c, void *d) {
    PHYSFS_getCdRomDirsCallback(c, d);
    return;
}

void _PHYSFS_getSearchPathCallback(PHYSFS_StringCallback c, void *d) {
    PHYSFS_getSearchPathCallback(c, d);
    return;
}

void _PHYSFS_enumerateFilesCallback(const char *dir, PHYSFS_EnumFilesCallback c, void *d) {
    PHYSFS_enumerateFilesCallback(dir, c, d);
    return;
}

void _PHYSFS_utf8FromUcs4(const PHYSFS_uint32 *src, char *dst, PHYSFS_uint64 len) {
    PHYSFS_utf8FromUcs4(src, dst, len);
    return;
}

void _PHYSFS_utf8ToUcs4(const char *src, PHYSFS_uint32 *dst, PHYSFS_uint64 len) {
    PHYSFS_utf8ToUcs4(src, dst, len);
    return;
}

void _PHYSFS_utf8FromUcs2(const PHYSFS_uint16 *src, char *dst, PHYSFS_uint64 len) {
    PHYSFS_utf8FromUcs2(src, dst, len);
    return;
}

void _PHYSFS_utf8ToUcs2(const char *src, PHYSFS_uint16 *dst, PHYSFS_uint64 len) {
    PHYSFS_utf8ToUcs2(src, dst, len);
    return;
}

void _PHYSFS_utf8FromLatin1(const char *src, char *dst, PHYSFS_uint64 len) {
    PHYSFS_utf8FromLatin1(src, dst, len);
    return;
}

int _PHYSFS_caseFold(const PHYSFS_uint32 from, PHYSFS_uint32 *to) {
    auto result = PHYSFS_caseFold(from, to);
    return result;
}

int _PHYSFS_utf8stricmp(const char *str1, const char *str2) {
    auto result = PHYSFS_utf8stricmp(str1, str2);
    return result;
}

int _PHYSFS_utf16stricmp(const PHYSFS_uint16 *str1, const PHYSFS_uint16 *str2) {
    auto result = PHYSFS_utf16stricmp(str1, str2);
    return result;
}

int _PHYSFS_ucs4stricmp(const PHYSFS_uint32 *str1, const PHYSFS_uint32 *str2) {
    auto result = PHYSFS_ucs4stricmp(str1, str2);
    return result;
}

int _PHYSFS_enumerate(const char *dir, PHYSFS_EnumerateCallback c, void *d) {
    auto result = PHYSFS_enumerate(dir, c, d);
    return result;
}

int _PHYSFS_unmount(const char *oldDir) {
    auto result = PHYSFS_unmount(oldDir);
    return result;
}

const PHYSFS_Allocator *_PHYSFS_getAllocator() {
    auto result = PHYSFS_getAllocator();
    return result;
}

int _PHYSFS_stat(const char *fname, PHYSFS_Stat *stat) {
    auto result = PHYSFS_stat(fname, stat);
    return result;
}

void _PHYSFS_utf8FromUtf16(const PHYSFS_uint16 *src, char *dst, PHYSFS_uint64 len) {
    PHYSFS_utf8FromUtf16(src, dst, len);
    return;
}

void _PHYSFS_utf8ToUtf16(const char *src, PHYSFS_uint16 *dst, PHYSFS_uint64 len) {
    PHYSFS_utf8ToUtf16(src, dst, len);
    return;
}

PHYSFS_sint64 _PHYSFS_readBytes(PHYSFS_File *handle, void *buffer, PHYSFS_uint64 len) {
    auto result = PHYSFS_readBytes(handle, buffer, len);
    return result;
}

PHYSFS_sint64 _PHYSFS_writeBytes(PHYSFS_File *handle, const void *buffer, PHYSFS_uint64 len) {
    auto result = PHYSFS_writeBytes(handle, buffer, len);
    return result;
}

int _PHYSFS_mountIo(PHYSFS_Io *io, const char *newDir, const char *mountPoint, int appendToPath) {
    auto result = PHYSFS_mountIo(io, newDir, mountPoint, appendToPath);
    return result;
}

int _PHYSFS_mountMemory(const void *buf, PHYSFS_uint64 len, void (*del)(void *), const char *newDir,
                        const char *mountPoint, int appendToPath) {
    auto result = PHYSFS_mountMemory(buf, len, del, newDir, mountPoint, appendToPath);
    return result;
}

int _PHYSFS_mountHandle(PHYSFS_File *file, const char *newDir, const char *mountPoint, int appendToPath) {
    auto result = PHYSFS_mountHandle(file, newDir, mountPoint, appendToPath);
    return result;
}

PHYSFS_ErrorCode _PHYSFS_getLastErrorCode() {
    auto result = PHYSFS_getLastErrorCode();
    return result;
}

const char *_PHYSFS_getErrorByCode(PHYSFS_ErrorCode code) {
    auto result = PHYSFS_getErrorByCode(code);
    return result;
}

void _PHYSFS_setErrorCode(PHYSFS_ErrorCode code) {
    PHYSFS_setErrorCode(code);
    return;
}

const char *_PHYSFS_getPrefDir(const char *org, const char *app) {
    auto result = PHYSFS_getPrefDir(org, app);
    return result;
}

int _PHYSFS_registerArchiver(const PHYSFS_Archiver *archiver) {
    auto result = PHYSFS_registerArchiver(archiver);
    return result;
}

int _PHYSFS_deregisterArchiver(const char *ext) {
    auto result = PHYSFS_deregisterArchiver(ext);
    return result;
}

int _PHYSFS_setRoot(const char *archive, const char *subdir) {
    auto result = PHYSFS_setRoot(archive, subdir);
    return result;
}
} // namespace stub
} // namespace physfs
} // namespace patchs
