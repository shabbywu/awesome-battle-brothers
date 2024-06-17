#ifndef BB_H
#define BB_H


#include <iostream>
// GLM
#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>


struct Rect {
    float left;
    float right;
    float bottom;
    float top;
};


struct FastDelegateClosure {
    void* m_pthis;
    void* m_pFunction;
};


struct FastDelegate {
    FastDelegateClosure m_Closure;
};


struct Ticket {
    void* pTag;
    uint32_t uMaxVertices;
    uint8_t bFlushImmediately;
    int32_t iZ;
    uint32_t uTextureID[0x8];
    void* pContext;
    FastDelegate onRender;
};


struct Font {
    void* vtable;
    void* m_pBuffer;
    std::string m_sName;
    int m_eStatus;
    uint8_t m_bIsExternal;
    uint32_t m_uReferenceCount;
    uint32_t m_uTextureID;
    // 字体图片的宽度
    int32_t m_iWidth;
    // 字体图片的高度
    int32_t m_iHeight;
    int32_t m_iChannels;
    uint32_t m_uSizeInMemory;
    int32_t m_iMaxMipmapLevel;
    int32_t m_iForceChannels;
    uint32_t m_uTextureFlags;
    int32_t m_iInternalFormat;
    uint8_t* m_pImageData;
    // 每个字符在 png 里的位置
    Rect m_Characters[0x100];
    // 每个字符的宽度, 用于计算边界和展示
    float m_fCharacterWidth[0x100];
    // 每个字符的高度, 20px/100px
    float m_fCharacterHeight;
};


struct FontHandler {
    void* vtable;
    Font* m_pPointee;
};


struct Brush {

};


struct BrushHandler {
    void* vtable;
    Brush* m_pPointee;
};


struct Color {
    union {
        uint32_t rgba;
        struct
        {
            uint8_t r;
            uint8_t g;
            uint8_t b;
            uint8_t a;
        } __inner1;
    } __inner0;
};


struct Text {
    void* vtable;
    uint32_t m_uReferenceCount;
    glm::vec2 m_vfPos;
    int32_t m_iZ;
    int32_t m_iLevel;
    float m_fRot;
    uint32_t m_uID;
    void* m_pChildren;
    void* m_pParent;
    void* m_pLayer;
    void* m_pLayerData;
    uint32_t m_uFlags;
    uint32_t m_uCustomFlags;
    Ticket* m_pTicket;
    FontHandler m_pFont;
    std::string m_sText;
    // 边界, top = m_fCharacterHeight * 行数(0xa换行), right = max(sum(m_fCharacterWidth))
    // 坐标原点在左下角
    Rect m_Bounds;
    // 坐标相关
    glm::vec2 m_vfPivot;
    // 坐标相关
    glm::vec2 m_vfOffset;
    BrushHandler m_pBackgroundBrush;
    float m_fScaleWithCameraFactor;
    // 字体大小
    float m_fSize;
    int m_eAlignmentHorizontal;
    int m_eAlignmentVertical;
    Color m_Color;
    Color m_BackgroundColor;
    uint8_t m_bIsBackgroundVisible;

    void __thiscall onRender(void* b, void* c) {
    }
};

#endif // !BB
