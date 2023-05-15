<script lang="ts">
import { Splitpanes, Pane } from 'splitpanes'

import { defineComponent, h } from 'vue'

export const Splitter = defineComponent({
  setup(props, { slots }) {
    return () => h(Splitpanes, props, slots)
  },
})

export const Panel = defineComponent({
  setup(props, { slots }) {
    return () => h(Pane, props, slots)
  },
})

export default {}
</script>

<style lang="scss">
@use '@/assets/themes/theme.scss';

@use 'splitpanes/dist/splitpanes.css';

.splitpanes__pane {
  display: flex;
  > * {
    flex: 1;
  }
}

.splitpanes--vertical > .splitpanes__splitter {
  width: 7px;
  cursor: col-resize;
  &:hover {
    background-color: theme.$backgroundColorHighlight;
  }
}

.splitpanes--horizontal > .splitpanes__splitter {
  height: 7px;
  cursor: row-resize;
  &:hover {
    background-color: theme.$backgroundColorHighlight;
  }
}

.splitpanes--horizontal .splitpanes__pane {
  transition: none !important;
}
</style>
