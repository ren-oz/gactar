<template>
  <b-tab-item class="code-tab" :label="framework">
    <div class="columns buttons">
      <div class="column">
        <strong>{{ defaultFileName }}.{{ fileExtension }}</strong> (generated
        code)
        <b-field class="is-pulled-right">
          <save-button
            :code="code"
            :default-name="defaultFileName"
            :file-extension="fileExtension"
          />
        </b-field>
      </div>
    </div>

    <code-mirror
      :key="count"
      :ref="refName"
      :amod-code="code"
      :mode="mode"
      :framework="framework"
      :read-only="true"
    />
  </b-tab-item>
</template>

<script lang="ts">
import Vue from 'vue'

import CodeMirror from './CodeMirror.vue'
import SaveButton from './SaveButton.vue'

interface Data {
  fileToLoad: string | null
  accept: string
  refName: string
  count: number
}

export default Vue.extend({
  components: { CodeMirror, SaveButton },

  props: {
    code: {
      type: String,
      required: true,
    },
    mode: {
      type: String,
      required: true,
    },
    fileExtension: {
      type: String,
      required: true,
    },
    framework: {
      type: String,
      required: true,
    },
    modelName: {
      type: String,
      required: true,
    },
  },

  data(): Data {
    return {
      fileToLoad: null,

      accept: '.' + this.mode + ',text/plain',
      refName: 'code-editor-' + this.mode,

      // This is used to prevent caching of the code-mirror data.
      // See https://stackoverflow.com/questions/48400302/vue-js-not-updating-props-in-child-when-parent-component-is-changing-the-propert
      count: 0,
    }
  },

  computed: {
    defaultFileName(): string {
      return this.framework + '_' + this.modelName
    },
  },

  watch: {
    code() {
      this.count += 1
    },
  },
})
</script>
