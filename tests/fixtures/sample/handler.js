import express from 'express';
import { UserService } from './services/user';

const router = express.Router();

router.get('/users', async (req, res) => {
  const users = await UserService.list();
  res.json(users);
});

router.post('/users', async (req, res) => {
  const user = await UserService.create(req.body);
  res.status(201).json(user);
});

export default router;
