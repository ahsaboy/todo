import { reactive, ref, type UnwrapNestedRefs } from 'vue'

export interface FormStateConfig<T extends Record<string, unknown>> {
  initialData: T
  onSubmit: (data: T) => Promise<void>
  validate?: (data: T) => string | null
}

export interface FormStateReturn<T extends Record<string, unknown>> {
  form: UnwrapNestedRefs<T>
  submitting: Ref<boolean>
  error: Ref<string>
  reset: () => void
  resetTo: (newData: T) => void
  handleSubmit: () => Promise<void>
}

export function useFormState<T extends Record<string, unknown>>(
  config: FormStateConfig<T>,
): FormStateReturn<T> {
  const form = reactive({ ...config.initialData }) as UnwrapNestedRefs<T>
  const submitting = ref(false)
  const error = ref('')

  function reset() {
    Object.assign(form, config.initialData)
    error.value = ''
  }

  function resetTo(newData: T) {
    Object.assign(form, newData)
    error.value = ''
  }

  async function handleSubmit() {
    error.value = ''
    if (config.validate) {
      const msg = config.validate(form)
      if (msg) { error.value = msg; return }
    }
    submitting.value = true
    try {
      await config.onSubmit(form)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '操作失败'
    } finally {
      submitting.value = false
    }
  }

  return { form, submitting, error, reset, resetTo, handleSubmit }
}
