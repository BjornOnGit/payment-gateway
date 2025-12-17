export type Settlement = {
  id: string
  merchant_account_id: string
  amount: number
  status: string
  external_reference?: string
  attempts: number
  created_at: string
  updated_at: string
}
