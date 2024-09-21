#pragma once
namespace sqext_buffer {
class Metadata {
  public:
    const int version = 1000;

  public:
    int getVersion() {
        return version;
    }

  public:
    static Metadata *getInstance() {
        static Metadata *instance;
        if (instance == nullptr)
            instance = new Metadata();
        return instance;
    }
};
} // namespace sqext_buffer
