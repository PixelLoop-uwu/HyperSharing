import { useState } from 'react'
import { FiEye, FiEyeOff } from 'react-icons/fi'
import GhostButton from '../actions/GhostButton'
import TextField, { type TextFieldProps } from '../forms/TextField'

/** TextField for passwords with a show/hide toggle. */
export default function PasswordField(props: Omit<TextFieldProps, 'type' | 'endSlot'>) {
  const [visible, setVisible] = useState(false)

  return (
    <TextField
      {...props}
      type={visible ? 'text' : 'password'}
      placeholder='Минимум 8 символов'
      endSlot={
        <GhostButton
          iconOnly
          aria-label={visible ? 'Скрыть пароль' : 'Показать пароль'}
          aria-pressed={visible}
          onClick={() => setVisible((v) => !v)}
          icon={visible ? <FiEyeOff className="size-4" /> : <FiEye className="size-4" />}
        />
      }
    />
  )
}
