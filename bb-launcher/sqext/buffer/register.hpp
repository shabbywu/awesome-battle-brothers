#pragma once
#include "../vm.hpp"
#include "metadata.hpp"
#include "reader.hpp"
#include "writer.hpp"
#include <sqbind17/sqbind17.hpp>

namespace sqext_buffer {
void static sqext_register_bufferio_impl(sqbind17::detail::VM &v) {
    vm = &v;

    sqbind17::detail::ClassDef<Metadata>(*vm, "Metadata").bindFunc(std::string("getVersion"), &Metadata::getVersion);

    sqbind17::detail::ClassDef<BufferReader>(*vm, "BufferReader")
        .bindFunc(std::string("getMetaData"), std::move(&BufferReader::getMetaData))
        .bindFunc(std::string("readBool"), std::move(&BufferReader::readBool))
        .bindFunc(std::string("readI8"), std::move(&BufferReader::readI8))
        .bindFunc(std::string("readU8"), std::move(&BufferReader::readU8))
        .bindFunc(std::string("readI16"), std::move(&BufferReader::readI16))
        .bindFunc(std::string("readU16"), std::move(&BufferReader::readU16))
        .bindFunc(std::string("readI32"), std::move(&BufferReader::readI32))
        .bindFunc(std::string("readU32"), std::move(&BufferReader::readU32))
        .bindFunc(std::string("readF32"), std::move(&BufferReader::readF32))
        .bindFunc(std::string("readString"), std::move(&BufferReader::readString))
        .bindFunc(std::string("loadBuffer"), std::move(&BufferReader::loadBuffer))
        .bindFunc(std::string("clear"), std::move(&BufferReader::clear));

    sqbind17::detail::ClassDef<BufferWriter>(*vm, "BufferWriter")
        .bindFunc(std::string("getMetaData"), std::move(&BufferWriter::getMetaData))
        .bindFunc(std::string("writeBool"), std::move(&BufferWriter::writeBool))
        .bindFunc(std::string("writeI8"), std::move(&BufferWriter::writeI8))
        .bindFunc(std::string("writeU8"), std::move(&BufferWriter::writeU8))
        .bindFunc(std::string("writeI16"), std::move(&BufferWriter::writeI16))
        .bindFunc(std::string("writeU16"), std::move(&BufferWriter::writeU16))
        .bindFunc(std::string("writeI32"), std::move(&BufferWriter::writeI32))
        .bindFunc(std::string("writeU32"), std::move(&BufferWriter::writeU32))
        .bindFunc(std::string("writeF32"), std::move(&BufferWriter::writeF32))
        .bindFunc(std::string("writeString"), std::move(&BufferWriter::writeString))
        .bindFunc(std::string("clear"), std::move(&BufferWriter::clear))
        .bindFunc(std::string("dumpBuffer"), std::move(&BufferWriter::dumpBuffer));

    auto roottable = sqbind17::detail::Table(_table(vm->roottable()), *vm);
    auto bufferio = sqbind17::detail::Table(*vm);

    bufferio.bindFunc(std::string("getSharedWriter"), &BufferWriter::getInstance);
    bufferio.bindFunc(std::string("createWriter"), []() -> BufferWriter * { return new BufferWriter(); });
    bufferio.bindFunc(std::string("getSharedReader"), &BufferReader::getInstance);
    bufferio.bindFunc(std::string("createReader"), []() -> BufferReader * { return new BufferReader(); });

    roottable.set(std::string("bufferio"), bufferio);
}
} // namespace sqext_buffer
