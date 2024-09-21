#pragma once
#include "metadata.hpp"
#include <base64.hpp>
#include <squirrel.h>
#include <string>
#include <vector>

namespace sqext_buffer {
class BufferReader {
  private:
    std::vector<unsigned char> buffer;
    int pos;

  private:
    template <typename T> T read() {
        auto size = sizeof(T);
        T *data = (T *)(buffer.data() + pos);
        pos += size;
        return *data;
    }

  public:
    BufferReader() : pos(0) {
    }
    BufferReader(std::vector<unsigned char> buffer) : buffer(buffer), pos(0) {};

  public:
    static BufferReader *getInstance() {
        static BufferReader *instance;
        if (instance == nullptr) {
            instance = new BufferReader();
        }
        return instance;
    }

  public:
    bool readBool() {
        return read<char>();
    }

    SQInteger readI8() {
        return read<char>();
    }

    SQUnsignedInteger readU8() {
        return read<unsigned char>();
    }

    SQInteger readI16() {
        return read<short>();
    }

    SQUnsignedInteger readU16() {
        return read<unsigned short>();
    }

    SQInteger readI32() {
        return read<int>();
    }

    SQUnsignedInteger readU32() {
        return read<unsigned int>();
    }

    SQFloat readF32() {
        return read<SQFloat>();
    }

    std::string readString() {
        auto size = readU32();
        std::string data(buffer.begin() + pos, buffer.begin() + pos + size);
        pos += size;
        return data;
    }

    void loadBuffer(std::string data) {
        buffer = base64::decode_into<std::vector<unsigned char>>(data);
        pos = 0;
    }

    void clear(){
        buffer.clear();
        pos = 0;
    }

  public:
    Metadata *getMetaData() {
        return Metadata::getInstance();
    }
};
} // namespace sqext_buffer
