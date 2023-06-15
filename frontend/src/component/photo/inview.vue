<template>
  <component
    :is="tag"
    v-observe-visibility="{
      callback: (isVisible) => {
        inView = isVisible
      },
      once: once,
      intersection: {
        rootMargin: rootMargin,
        threshold: threshold
      }
    }"
  >
    <slot :visible="inView" />
  </component>
</template>

<script>
export default {
  name: 'InView',
  props: {
    tag: {
      type: String,
      default: 'div'
    },
    once: {
      type: Boolean,
      default: false
    },
    threshold: {
      type: Number,
      default: 0
    },
    rootMargin: {
      type: String,
      default: '0px 0px 0px 0px'
    }
  },
  data() {
    return {
      inView: false
    }
  }
}
</script>