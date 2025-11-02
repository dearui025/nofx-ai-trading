// Supabase client configuration
import { createClient } from '@supabase/supabase-js';

const supabaseUrl = "https://eqzurdzoaxibothslnna.supabase.co";
const supabaseAnonKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImVxenVyZHpvYXhpYm90aHNsbm5hIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjE4NzY2NjUsImV4cCI6MjA3NzQ1MjY2NX0.h2EQOkofLavh-DL68AGfFX7ZvJ4SipNsiO7K5uTh20Y";

export const supabase = createClient(supabaseUrl, supabaseAnonKey);
