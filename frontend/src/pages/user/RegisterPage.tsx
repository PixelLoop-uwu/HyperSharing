import { motion } from 'framer-motion'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AuthShell, useRegister } from '@/features/user'
import { rise, stagger } from '@/shared/lib/motion'
import { useSubmit } from '@/shared/lib/useSubmit'
import FormError from '@/shared/ui/forms/FormError'
import PasswordField from '@/shared/ui/forms/PasswordField'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'
import TextField from '@/shared/ui/forms/TextField'
import GhostLink from '@/shared/ui/actions/GhostLink'

export default function RegisterPage() {
  const register = useRegister()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  const { onSubmit, error, pending } = useSubmit(async () => {
    await register.mutateAsync({ email, username, password })
    navigate('/')
  }, 'Не удалось зарегистрироваться')

  return (
    <AuthShell
      title="Регистрация"
      lead="Аккаунт позволяет управлять своими публикациями."
      footer={
        <>
          Уже есть аккаунт? <GhostLink to="/login" accentTextOnly>Войти</GhostLink>
        </>
      }
    >
      <motion.form variants={stagger} onSubmit={onSubmit} className="mt-6 space-y-4">
        <motion.div variants={rise}>
          <TextField
            label="Email"
            type="email"
            required
            autoFocus
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="goida@example.com"
          />
        </motion.div>
        <motion.div variants={rise}>
          <TextField
            label="Имя пользователя"
            type="text"
            required
            minLength={3}
            maxLength={32}
            pattern="[a-zA-Z0-9_]+"
            title="Латиница, цифры и нижнее подчёркивание"
            autoComplete="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="anytthing"
            hint="От 3 до 32 символов: латиница, цифры и _"
          />
        </motion.div>
        <motion.div variants={rise}>
          <PasswordField
            label="Пароль"
            required
            minLength={8}
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </motion.div>
        <FormError message={error} />
        <motion.div variants={rise}>
          <PrimaryButton type="submit" pending={pending} className="w-full justify-center mt-4">
            Создать аккаунт
          </PrimaryButton>
        </motion.div>
      </motion.form>
    </AuthShell>
  )
}
