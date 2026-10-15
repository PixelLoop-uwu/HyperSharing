import { motion } from 'framer-motion'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import AuthShell from '@/features/user/ui/AuthShell'
import { useAuth } from '@/features/user/store'
import { rise, stagger } from '@/shared/lib/motion'
import { useSubmit } from '@/shared/lib/useSubmit'
import FormError from '@/shared/ui/forms/FormError'
import PasswordField from '@/shared/ui/forms/PasswordField'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'
import TextField from '@/shared/ui/forms/TextField'
import TextLink from '@/shared/ui/actions/TextLink'

export default function LoginPage() {
  const login = useAuth((s) => s.login)
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const { onSubmit, error, pending } = useSubmit(async () => {
    await login({ email, password })
    navigate('/')
  }, 'Не удалось войти')

  return (
    <AuthShell
      title="Вход"
      lead="Войдите в аккаунт, чтобы увеличить лимиты."
      footer={
        <>
          Нет аккаунта? <TextLink to="/register">Регистрация</TextLink>
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
            placeholder="you@example.com"
          />
        </motion.div>
        <motion.div variants={rise}>
          <PasswordField
            label="Пароль"
            required
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </motion.div>
        <FormError message={error} />
        <motion.div variants={rise}>
          <PrimaryButton type="submit" pending={pending} className="w-full justify-center mt-4">
            Войти
          </PrimaryButton>
        </motion.div>
      </motion.form>
    </AuthShell>
  )
}
