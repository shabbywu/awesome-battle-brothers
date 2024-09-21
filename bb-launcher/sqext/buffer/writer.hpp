#pragma once
#include "metadata.hpp"
#include <base64.hpp>
#include <squirrel.h>
#include <string>
#include <vector>

namespace sqext_buffer {
class BufferWriter {
  private:
    std::vector<unsigned char> buffer;
    int pos;

  private:
    template <typename T> auto write(T data) {
        auto size = sizeof(T);
        for (int i = 0; i < size; i++) {
            unsigned char part = ((unsigned char *)&data)[i];
            buffer.push_back(part);
        }
        pos += size;
        return size;
    }

  public:
    BufferWriter() : pos(0) {
    }
    BufferWriter(std::vector<unsigned char> buffer) : buffer(buffer), pos(buffer.size()) {};

  public:
    static BufferWriter *getInstance() {
        static BufferWriter *instance;
        if (instance == nullptr) {
            instance = new BufferWriter();
        }
        return instance;
    }

  public:
    auto writeBool(bool data) {
        return write((char)(data));
    }

    auto writeI8(SQInteger data) {
        return write((char)(data));
    }

    auto writeU8(SQInteger data) {
        return write((unsigned char)(data));
    }

    auto writeI16(SQInteger data) {
        return write((short)(data));
    }

    auto writeU16(SQInteger data) {
        return write((unsigned short)(data));
    }

    auto writeI32(SQInteger data) {
        return write((int)(data));
    }

    auto writeU32(SQInteger data) {
        return write((unsigned int)(data));
    }

    auto writeF32(SQFloat data) {
        return write((SQFloat)(data));
    }

    auto writeString(std::string data) {
        auto header_size = writeU32(data.size());
        auto c_str = data.data();
        for (auto i = 0; i < data.size(); i++) {
            unsigned char part = *(c_str + i);
            buffer.push_back(part);
        }
        pos += data.size();
        return header_size + data.size();
    }

    auto dumpBuffer() {
        return base64::encode_into<std::string>(buffer.begin(), buffer.end());
    }

    void clear() {
        buffer.clear();
        pos = 0;
    }

  public:
    Metadata *getMetaData() {
        return Metadata::getInstance();
    }
};
} // namespace sqext_buffer
