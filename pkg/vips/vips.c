#include "vips.h"

void remove_metadata(VipsImage *in);

int thumbnail(const char *filename, const char *outputname, int width, int height, int crop, int q)
{
    int ret;
    VipsImage *image;

    if (crop == -1)
    {
        ret = vips_thumbnail(filename, &image, width, NULL);
    }
    else
    {
        ret = vips_thumbnail(filename, &image, width, "height", height, "crop", crop, NULL);
    }

    if (ret)
    {
        return -1;
    }

    remove_metadata(image);

    if (vips_image_write_to_file(image, outputname, "Q", q, NULL))
    {
        VIPS_UNREF(image);
        return (-1);
    }
    VIPS_UNREF(image);
    return 0;
}

// Keeps ICC profile, orientation and pages metadata
void remove_metadata(VipsImage *in) {
  gchar **fields = vips_image_get_fields(in);

  for (int i = 0; fields[i] != NULL; i++) {
    if (strncmp(fields[i], VIPS_META_ICC_NAME, sizeof(VIPS_META_ICC_NAME)) &&
        strncmp(fields[i], VIPS_META_ORIENTATION, sizeof(VIPS_META_ORIENTATION)) &&
        strncmp(fields[i], VIPS_META_N_PAGES, sizeof(VIPS_META_N_PAGES)) &&
        strncmp(fields[i], VIPS_META_PAGE_HEIGHT, sizeof(VIPS_META_PAGE_HEIGHT))
        ) {
      vips_image_remove(in, fields[i]);
    }
  }

  g_strfreev(fields);
}
